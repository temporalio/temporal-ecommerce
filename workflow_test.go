package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
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
	}, time.Millisecond*1)

	s.env.RegisterDelayedCallback(func() {
		res, err := s.env.QueryWorkflow("getCart")
		s.NoError(err)
		err = res.Get(&cart)
		s.NoError(err)
		s.Equal(1, len(cart.Items))
	}, time.Millisecond*2)

	s.env.ExecuteWorkflow(CartWorkflow, cart)

	s.True(s.env.IsWorkflowCompleted())
}

func (s *UnitTestSuite) Test_RemoveFromCart() {
	cart := CartState{Items: make([]CartItem, 0)}

	// Add 2 items to the cart
	s.env.RegisterDelayedCallback(func() {
		update := AddToCartSignal{
			Route: RouteTypes.ADD_TO_CART,
			Item: CartItem{ProductId: 1, Quantity: 2},
		}
		s.env.SignalWorkflow("cartMessages", update)
	}, time.Millisecond*1)

	// Remove 1 item from the cart
	s.env.RegisterDelayedCallback(func() {
		res, err := s.env.QueryWorkflow("getCart")
		s.NoError(err)
		err = res.Get(&cart)
		s.NoError(err)
		s.Equal(len(cart.Items), 1)
		s.Equal(cart.Items[0].Quantity, 2)

		update := AddToCartSignal{
			Route: RouteTypes.REMOVE_FROM_CART,
			Item: CartItem{ProductId: 1, Quantity: 1},
		}
		s.env.SignalWorkflow("cartMessages", update)
	}, time.Millisecond*2)

	s.env.ExecuteWorkflow(CartWorkflow, cart)

	s.True(s.env.IsWorkflowCompleted())

	res, err := s.env.QueryWorkflow("getCart")
	s.NoError(err)
	err = res.Get(&cart)
	s.NoError(err)
	s.Equal(1, len(cart.Items))
	s.Equal(cart.Items[0].Quantity, 1)
}

func (s *UnitTestSuite) Test_Checkout() {
	cart := CartState{Items: make([]CartItem, 0)}

	var a *Activities

	s.env.OnActivity(a.CreateStripeCharge, mock.Anything, mock.Anything).Return(
		func(_ context.Context, _ CartState) (error) {
			return nil
		})

	// Add a product to the cart
	s.env.RegisterDelayedCallback(func() {
		update := AddToCartSignal{
			Route: RouteTypes.ADD_TO_CART,
			Item: CartItem{ProductId: 1, Quantity: 1},
		}
		s.env.SignalWorkflow("cartMessages", update)
	}, time.Millisecond*1)

	// Check out
	s.env.RegisterDelayedCallback(func() {
		res, err := s.env.QueryWorkflow("getCart")
		s.NoError(err)
		err = res.Get(&cart)
		s.NoError(err)
		s.Equal(len(cart.Items), 1)
		s.Equal(cart.Items[0].Quantity, 1)

		update := CheckoutSignal{
			Route: RouteTypes.CHECKOUT,
			Email: "test@temporal.io",
		}
		s.env.SignalWorkflow("cartMessages", update)
	}, time.Millisecond*2)

	// Workflow should be completed after checking out
	s.env.RegisterDelayedCallback(func() {
		s.True(s.env.IsWorkflowCompleted())
	}, time.Millisecond*3)

	s.env.ExecuteWorkflow(CartWorkflow, cart)
}

func (s *UnitTestSuite) Test_AbandonedCart() {
	cart := CartState{Items: make([]CartItem, 0)}

	var a *Activities

	sendTo := ""
	s.env.OnActivity(a.SendAbandonedCartEmail, mock.Anything, mock.Anything).Return(
		func(_ context.Context, _sendTo string) (error) {
			sendTo = _sendTo
			return nil
		})

	// Add a product to the cart
	s.env.RegisterDelayedCallback(func() {
		update := AddToCartSignal{
			Route: RouteTypes.ADD_TO_CART,
			Item: CartItem{ProductId: 1, Quantity: 1},
		}
		s.env.SignalWorkflow("cartMessages", update)

		updateEmail := UpdateEmailSignal{
			Route: RouteTypes.UPDATE_EMAIL,
			Email: "abandoned_test@temporal.io",
		}
		s.env.SignalWorkflow("cartMessages", updateEmail)
	}, time.Millisecond*1)

	// Wait for 10 mins and make sure abandoned cart email has been sent
	s.env.RegisterDelayedCallback(func() {
		s.Equal(sendTo, "abandoned_test@temporal.io")
	}, abandonedCartTimeout + time.Millisecond*2)

	s.env.ExecuteWorkflow(CartWorkflow, cart)

	s.True(s.env.IsWorkflowCompleted())
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}