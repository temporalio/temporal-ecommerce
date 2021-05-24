package app

import (
	//"errors"
	"testing"
	"fmt"

	//"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	//"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	// "go.temporal.io/sdk/client"

	"time"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_AddToCart() {
	cart := CartState{Items: make([]CartItem, 0)}

	s.env.RegisterDelayedCallback(func() {
		// suite.go:63: test panicked: runtime error: invalid memory address or nil pointer dereference
		/*res, err := s.env.QueryWorkflow("getCart")
		s.NoError(err)
		err = res.Get(&cart)
		s.NoError(err)
		s.Equal(len(cart.Items), 0)*/

		update := AddToCartSignal{
			Route: RouteTypes.ADD_TO_CART,
			Item: CartItem{ProductId: 1, Quantity: 1},
		}
		s.env.SignalWorkflow("cartMessages", update)
		fmt.Println(cart.Items)
	}, time.Millisecond*0)

	s.env.ExecuteWorkflow(CartWorkflow, cart)

	res, err := s.env.QueryWorkflow("getCart")
	s.NoError(err)
	err = res.Get(&cart)
	s.NoError(err)
	s.Equal(1, len(cart.Items))

	s.True(s.env.IsWorkflowCompleted())
}

func (s *UnitTestSuite) Test_RemoveFromCart() {
	cart := CartState{Items: make([]CartItem, 0)}

	s.env.RegisterDelayedCallback(func() {
		update := AddToCartSignal{
			Route: RouteTypes.ADD_TO_CART,
			Item: CartItem{ProductId: 1, Quantity: 2},
		}
		s.env.SignalWorkflow("cartMessages", update)

		update = AddToCartSignal{
			Route: RouteTypes.REMOVE_FROM_CART,
			Item: CartItem{ProductId: 1, Quantity: 1},
		}
		s.env.SignalWorkflow("cartMessages", update)
	}, time.Millisecond*0)

	s.env.ExecuteWorkflow(CartWorkflow, cart)

	res, err := s.env.QueryWorkflow("getCart")
	s.NoError(err)
	err = res.Get(&cart)
	s.NoError(err)
	s.Equal(1, len(cart.Items))

	s.True(s.env.IsWorkflowCompleted())
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}