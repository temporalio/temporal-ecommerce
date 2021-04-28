Temporal promises to help you build invincible apps.
To make this possible, Temporal introduces new design patterns that are very different from the traditional web app architecture.

Instead of API endpoints that talk to a database over the network, your API endpoints instead call in-memory _Workflows_ that store state internally.
Temporal handles persisting the state of your Workflows and distributing your Workflow between Workers as necessary.
You as the developer are responsible for implementing workflows and activities as normal Go code, Temporal handles the data persistence and horizontal scaling for you.

In this blog post, I'll show how to build a shopping cart using long-lived Workflows.
You can find the [full source code for this shopping cart on GitHub](https://github.com/vkarpov15/temporal-ecommerce).

## Shopping Cart Workflow

In traditional web app architecture, a user's shopping cart is stored as a row or document in a database.
While you can store shopping carts in a separate database using Temporal, you have another option: you can represent a shopping cart as a long-lived workflow.

A workflow is a Go function that takes 2 parameters: a Temporal workflow context `ctx`, and an arbitrary `value`. 
A workflow can run for an arbitrarily long period of time, and Temporal can handle pausing and restarting the workflow.
A workflow can then share its state via _queries_ and modify its state in response to _signals_.

Below is a simplified shopping cart that adds a new product to the cart every time it receives a `updateCart` signal.

```go
package app

import (
	"go.temporal.io/sdk/workflow"
)

type (
	CartItem struct {
		ProductId int
		Quantity  int
	}

	CartState struct {
		Items []CartItem
		Email string
	}
)

func CartWorkflowExample(ctx workflow.Context, state CartState) error {
	logger := workflow.GetLogger(ctx)

	err := workflow.SetQueryHandler(ctx, "getCart", func(input []byte) (CartState, error) {
		return state, nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed.", "Error", err)
		return err
	}

	channel := workflow.GetSignalChannel(ctx, "cartMessages")
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(channel, func(c workflow.ReceiveChannel, _ bool) {
		var signal interface{}
		c.Receive(ctx, &signal)
		state.Items = append(state.Items, CartItem{ProductId: 0, Quantity: 1})
	})

	for {
		selector.Select(ctx)
	}

	return nil
}
```

To run a workflow, you need to create a worker process.
A Temporal _worker_ listens for events on a queue and has a list of registered workflows that it can run in response to messages on the queue.
Below is the largely boilerplate `worker/main.go` file:

```go
package main

import (
	"log"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"

	"temporal-ecommerce/app"
)

func main() {
	// Create the client object just once per process
	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()
	// This worker hosts both Worker and Activity functions
	w := worker.New(c, "CART_TASK_QUEUE", worker.Options{})

	w.RegisterWorkflow(app.CartWorkflowExample)
	// Start listening to the Task Queue
	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalln("unable to start Worker", err)
	}
}
```

In order to see this shopping cart workflow in action, you can create a _starter_ that sends queries and signals
to modify the shopping cart.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"temporal-ecommerce/app"

	"go.temporal.io/sdk/client"
)

func main() {
	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalln("unable to create Temporal client", err)
	}
	defer c.Close()

	workflowID := "CART-" + fmt.Sprintf("%d", time.Now().Unix())

	options := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "CART_TASK_QUEUE_2",
	}

	state := app.CartState{Items: make([]app.CartItem, 0)}
	we, err := c.ExecuteWorkflow(context.Background(), options, app.CartWorkflowExample, state)
	if err != nil {
		log.Fatalln("unable to execute workflow", err)
	}

	err = c.SignalWorkflow(context.Background(), workflowID, we.GetRunID(), "cartMessages", nil)

	resp, err := c.QueryWorkflow(context.Background(), workflowID, we.GetRunID(), "getCart")
	if err != nil {
		log.Fatalln("Unable to query workflow", err)
	}
	var result interface{}
	if err := resp.Get(&result); err != nil {
		log.Fatalln("Unable to decode query result", err)
	}
	// Prints a message similar to:
	// 2021/03/31 15:43:54 Received query result Result map[Email: Items:[map[ProductId:0 Quantity:1]]]
	log.Println("Received query result", "Result", result)
}
```

## Adding and Removing Elements from the Cart

In order to support adding and removing elements from the cart, the workflow needs to respond to different types of signals.
Signals let you send arbitrary signals to workflows.
The below code listens to a signal channel for messages that either add or remove items from a shopping cart.

```golang
channel := workflow.GetSignalChannel(ctx, "cartMessages")
selector := workflow.NewSelector(ctx)

selector.AddReceive(channel, func(c workflow.ReceiveChannel, _ bool) {
	var signal interface{}
	c.Receive(ctx, &signal)

	var routeSignal RouteSignal
	err := mapstructure.Decode(signal, &routeSignal)
	if err != nil {
		logger.Error("Invalid signal type %v", err)
		return
	}

	switch {
	case routeSignal.Route == RouteTypes.ADD_TO_CART:
		var message AddToCartSignal
		err := mapstructure.Decode(signal, &message)
		if err != nil {
			logger.Error("Invalid signal type %v", err)
			return
		}

		AddToCart(&state, message.Item)
	case routeSignal.Route == RouteTypes.REMOVE_FROM_CART:
		var message RemoveFromCartSignal
		err := mapstructure.Decode(signal, &message)
		if err != nil {
			logger.Error("Invalid signal type %v", err)
			return
		}

		RemoveFromCart(&state, message.Item)
})

for {
	selector.Select(ctx)
}
```

All the `AddToCart()` functions and `RemoveFromCart()` functions need to do is modify the `state.Items` array.
Temporal is responsible for persisting and distributing the `state`.

```golang
func AddToCart(state *CartState, item CartItem) {
	for i := range state.Items {
		if state.Items[i].ProductId != item.ProductId {
			continue
		}

		state.Items[i].Quantity += item.Quantity
		return
	}

	state.Items = append(state.Items, item)
}

func RemoveFromCart(state *CartState, item CartItem) {
	for i := range state.Items {
		if state.Items[i].ProductId != item.ProductId {
			continue
		}

		state.Items[i].Quantity -= item.Quantity
		if state.Items[i].Quantity <= 0 {
			state.Items = append(state.Items[:i], state.Items[i+1:]...)
		}
		break
	}
}
```

## Next Up

Temporal introduces a new way of building web applications: instead of storing a shopping cart in a database, you
can represent a shopping cart as a long-lived Workflow.
For simple CRUD applications like this shopping cart app so far, this pattern doesn't make things significantly easier.
Next up, we'll look at a case where Temporal's long-lived Workflows shine: sending a reminder email if the user abandons their cart.
