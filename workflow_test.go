package app

import (
	//"errors"
	"testing"

	//"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	//"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"

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
	s.env.ExecuteWorkflow(CartWorkflow, cart)

	res, err := s.env.QueryWorkflow("getCart")
	s.NoError(err)
	err = res.Get(&cart)
	s.NoError(err)
	s.Equal(len(cart.Items), 0)

	s.env.RegisterDelayedCallback(func() {
		update := AddToCartSignal{
			Route: RouteTypes.ADD_TO_CART,
			Item: CartItem{ProductId: 1, Quantity: 1},
		}
		// Doesn't execute
		s.env.SignalWorkflow("cartMessages", update)
	
		res, err = s.env.QueryWorkflow("getCart")
		s.NoError(err)
		err = res.Get(&cart)
		s.NoError(err)
		// expected: 1, actual: 0
		s.Equal(1, len(cart.Items))
	}, time.Millisecond*0)
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}