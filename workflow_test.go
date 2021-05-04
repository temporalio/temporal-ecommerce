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

	// From https://pkg.go.dev/go.temporal.io/sdk@v1.6.0/internal#TestWorkflowEnvironment.SetStartWorkflowOptions
	/* s.env.SetStartWorkflowOptions(client.StartWorkflowOptions{
		WorkflowExecutionTimeout: 10*time.Minute,
	}) */
	// s.env.SetTestTimeout(1 * time.Minute)

	s.env.RegisterDelayedCallback(func() {
		fmt.Println("delayed")
		res, err := s.env.QueryWorkflow("getCart")
		s.NoError(err)
		err = res.Get(&cart)
		s.NoError(err)
		s.Equal(len(cart.Items), 0)

		update := AddToCartSignal{
			Route: RouteTypes.ADD_TO_CART,
			Item: CartItem{ProductId: 1, Quantity: 1},
		}
		s.env.SignalWorkflow("cartMessages", update)
	
		res, err = s.env.QueryWorkflow("getCart")
		s.NoError(err)
		err = res.Get(&cart)
		s.NoError(err)
		s.Equal(2, len(cart.Items))
	}, time.Millisecond*0)

	s.env.ExecuteWorkflow(CartWorkflow, cart)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}