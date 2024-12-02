# temporal-ecommerce

This is a demo app for a tutorial showing the process of developing a [Temporal eCommerce application in Go](https://learn.temporal.io/tutorials/go/build-an-ecommerce-app/), using the Stripe and Mailgun APIs.

## Instructions

To run the worker and server, you must set the `STRIPE_PRIVATE_KEY`, `MAILGUN_DOMAIN`, and `MAILGUN_PRIVATE_KEY` environment variables.
You can set the values to "test", which will allow you to add and remove elements from your cart.
But you won't be able to checkout or receive abandoned cart notifications if these values aren't set.

To run the worker, make sure you have a local instance of Temporal Server running (e.g. with [the Temporal CLI](https://github.com/temporalio/cli)), then run:

```bash
env STRIPE_PRIVATE_KEY=stripe-key-here env MAILGUN_DOMAIN=mailgun-domain-here env MAILGUN_PRIVATE_KEY=mailgun-private-key-here go run worker/main.go
```

To run the API server, you must also set the `PORT` environment variable as follows.

```bash
env STRIPE_PRIVATE_KEY=stripe-key-here env MAILGUN_DOMAIN=mailgun-domain-here env MAILGUN_PRIVATE_KEY=mailgun-private-key-here env PORT=3001 go run api/main.go
```

You can then run the UI on port 8080:

```
cd frontend
npm install
npm start
```

## Interacting with the API server with cURL

Here is a guide to the basic routes that you can see and what they expect:

```bash
# get items
curl http://localhost:3001/products

# response:
# {"products":[
    # {"Id":0,"Name":"iPhone 12 Pro","Description":"Test","Image":"https://images.unsplash.com/photo-1603921326210-6edd2d60ca68","Price":999},
    # {"Id":1,"Name":"iPhone 12","Description":"Test","Image":"https://images.unsplash.com/photo-1611472173362-3f53dbd65d80","Price":699},
    # {"Id":2,"Name":"iPhone SE","Description":"399","Image":"https://images.unsplash.com/photo-1529618160092-2f8ccc8e087b","Price":399},
    # {"Id":3,"Name":"iPhone 11","Description":"599","Image":"https://images.unsplash.com/photo-1574755393849-623942496936","Price":599}
# ]}

# create cart
curl -X POST http://localhost:3001/cart

# response:
# {"cart":{"Items":[],"Email":""},
#  "workflowID":"CART-1619483151"}

# add item
curl -X PUT -d '{"ProductId":3,"Quantity":1}' -H 'Content-Type: application/json' http://localhost:3001/cart/CART-1619483151/4a4436be-3307-42ea-a9ab-3b63f5520bee/add

# response: {"ok":1}

# get cart
curl http://localhost:3001/cart/CART-1619483151/4a4436be-3307-42ea-a9ab-3b63f5520bee

# response:
# {"Email":"","Items":[{"ProductId":3,"Quantity":1}]}
```

## Interacting with the API server with Node.js

Below is a Node.js script that creates a new cart, adds/removes some items, and checks out.

```javascript
'use strict';

const assert = require('assert');
const axios = require('axios');

void async function main() {
  let { data } = await axios.post('http://localhost:3001/cart');

  const { workflowID } = data;
  console.log(workflowID)

  await axios.put(`http://localhost:3001/cart/${workflowID}/add`, { ProductID: 1, Quantity: 2 });

  ({ data } = await axios.get(`http://localhost:3001/cart/${workflowID}`));
  console.log(data);
  assert.deepEqual(data.Items, [ { ProductId: 1, Quantity: 2 } ]);

  await axios.put(`http://localhost:3001/cart/${workflowID}/remove`, { ProductID: 1, Quantity: 1 });

  ({ data } = await axios.get(`http://localhost:3001/cart/${workflowID}`));
  console.log(data);
  assert.deepEqual(data.Items, [ { ProductId: 1, Quantity: 1 } ]);

  await axios.put(`http://localhost:3001/cart/${workflowID}/checkout`, { Email: 'val@temporal.io' });

  ({ data } = await axios.get(`http://localhost:3001/cart/${workflowID}`));
  console.log(data);
}();
```

## Notes on Testing

The following is a basic setup for testing a Temporal Workflow using `go test` and [Testify](https://github.com/stretchr/testify) based on [Temporal's Go testing docs](https://docs.temporal.io/dev-guide/go/testing).

You can find the full source code for the test suite in the `workflow_test.go` file.

```go
package app

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/client"

	"time"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *UnitTestSuite) SetupTest() {
	// You are responsible for calling `NewTestWorkflowEnvironment()` to initialize
	// Temporal's testing utilities, but you can also add any other setup you need
	// in this function.
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *UnitTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}
```

The most important property is the `env` property, which is an instance of [Temporal's `TestWorkflowEnvironment` struct](https://pkg.go.dev/go.temporal.io/temporal/internal#TestWorkflowEnvironment).
A `TestWorkflowEnvironment` provides utilities for testing Workflows, including executing Workflows, mocking Activities, and Signaling and Querying test Workflows.

The [testify package](https://github.com/stretchr/testify) also provides utilities for organizing tests, including setting up and tearing down test suites using `SetupTest()` and `AfterTest()`.
For example, you can define multiple test suites as shown below.

```go
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

type IntegrationTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite

	env *testsuite.TestWorkflowEnvironment
}

func (s *IntegrationTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
}

func (s *IntegrationTestSuite) AfterTest(suiteName, testName string) {
	s.env.AssertExpectations(s.T())
}
```

### Querying Workflows in tests

Remember that, in this app, a shopping cart is a Workflow.
To get the current state of the shopping cart, you send a Query to the Workflow, and to update the cart you send a Signal to the Workflow.

The below code shows how you can use `env.QueryWorkflow()` to send a Query to the shopping cart Workflow.

```go
func (s *UnitTestSuite) Test_QueryCart() {
	cart := CartState{Items: make([]CartItem, 0)}

	s.env.ExecuteWorkflow(CartWorkflow, cart)

  // Note that `ExecuteWorkflow()` is blocking: the Workflow is done by the time
  // the test gets to this line.
	s.True(s.env.IsWorkflowCompleted())

  // Send a query to the Workflow and assert that the shopping cart is still empty
	res, err := s.env.QueryWorkflow("getCart")
	s.NoError(err)
	err = res.Get(&cart)
	s.NoError(err)
	s.Equal(0, len(cart.Items))
}
```

Note that the above code Queries the Workflow _after the Workflow is done_.
In order to interact with the Workflow via Queries and Signals while the Workflow is running, you should use the test environment's `RegisterDelayedCallback()` function as shown below.
Make sure you call `RegisterDelayedCallback()` _before_ `ExecuteWorkflow()`, otherwise Temporal will execute the entire Workflow without executing the callback.

```go
func (s *UnitTestSuite) Test_IntermediateQuery() {
	cart := CartState{Items: make([]CartItem, 0)}

  // Register a callback to execute after 1 millisecond elapses in the Workflow.
	s.env.RegisterDelayedCallback(func() {
		res, err := s.env.QueryWorkflow("getCart")
		s.NoError(err)
		err = res.Get(&cart)
		s.NoError(err)
		s.Equal(len(cart.Items), 0)
	}, time.Millisecond*1)

	s.env.ExecuteWorkflow(CartWorkflow, cart)

	s.True(s.env.IsWorkflowCompleted())
}
```

You can Query a Workflow after it is completed, but you can't Signal a Workflow after it is completed.

### Signaling Workflows in tests

So in order to Signal a Workflow from your tests, you need to use `RegisterDelayedCallback()`.
Just remember that Signaling is asynchronous, so you need to add a separate `RegisterDelayedCallback()` to read the result of your Signal using a Query.
For example, below is a test case for the `AddToCart()` method.

```go
func (s *UnitTestSuite) Test_AddToCart() {
	cart := CartState{Items: make([]CartItem, 0)}

  // First callback at 1ms: query to make sure the cart is empty, and signal to add an item.
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

  // Second callback at 2ms: query to make sure the item is in the cart
  // This needs to be a separate callback, `s.Equal(1, len(cart.Items))` would
  // fail if it were in the 1ms callback.
	s.env.RegisterDelayedCallback(func() {
		res, err := s.env.QueryWorkflow("getCart")
		s.NoError(err)
		err = res.Get(&cart)
		s.NoError(err)

    s.Equal(1, len(cart.Items))
    s.Equal(1, cart.Items[0].Quantity)
	}, time.Millisecond*2)

	s.env.ExecuteWorkflow(CartWorkflow, cart)

	s.True(s.env.IsWorkflowCompleted())
}
```

### Sending multiple Signals to Workflows in tests

Similarly, if you want to send a Query to check the state of the Workflow between Signals, you should put the Query in a separate `RegisterDelayedCallback()` call.
You can move any Queries that don't have any Signals after them to after the `ExecuteWorkflow()` call.
For example, below is a test case for the `RemoveFromCart()` method.

```go
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

	// Query the current state and then remove 1 item from the cart
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

	// Since there's no more Signals, no need to put this Query in a
	// `RegisterDelayedCallback()` call.
	res, err := s.env.QueryWorkflow("getCart")
	s.NoError(err)
	err = res.Get(&cart)
	s.NoError(err)
	s.Equal(1, len(cart.Items))
	s.Equal(cart.Items[0].Quantity, 1)
}
```

This covers testing the basic functionality of adding items to and removing items from the shopping cart.
But what about testing more sophisticated features, like testing that the Workflow sends an abandoned cart email after 10 minutes?

### Mocking Activities and Controlling Time

Temporal's test environment makes it easy to mock Activities, replacing them with a stubbed out function.
For example, the below test asserts that sending a checkout Signal calls the `CreateStripeCharge` Activity with the correct receipt email using the test environment's `OnActivity()` function.

```go
func (s *UnitTestSuite) Test_Checkout() {
	cart := CartState{Items: make([]CartItem, 0)}

	var a *Activities
	sendTo := ""

	s.env.OnActivity(a.CreateStripeCharge, mock.Anything, mock.Anything).Return(
		func(_ context.Context, cart CartState) (error) {
			sendTo = cart.Email
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

	s.Equal(sendTo, "test@temporal.io")
}
```

What about testing the abandoned cart email?
Normally, testing the abandoned cart email is tricky because it involves waiting for 10 minutes.
The key insight is that Temporal's test environment advances time internally, and time in the test environment is **not** [wall-clock time](https://en.wikipedia.org/wiki/Elapsed_real_time).

The `RegisterDelayedCallback()` function ties into the test environment's internal notion of time.
Calling `RegisterDelayedCallback(fn, time.Minute*5)` does **not** tell the test environment to wait for 5 minutes of wall-clock time.
That means testing the abandoned cart email is easy: mock out the `SendAbandonedCartEmail()` activity and use `RegisterDelayedCallback()` with the `abandonedCartTimeout` as shown below.

```go
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

	// Wait for 10 mins and make sure abandoned cart email has been sent. The extra
	// 2ms is because signals are async, so the last change to the cart happens at 2ms.
	s.env.RegisterDelayedCallback(func() {
		s.Equal(sendTo, "abandoned_test@temporal.io")
	}, abandonedCartTimeout + time.Millisecond*2)

	s.env.ExecuteWorkflow(CartWorkflow, cart)

	s.True(s.env.IsWorkflowCompleted())
}
```
