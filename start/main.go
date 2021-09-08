// @@@SNIPSTART temporal-ecommerce-starter
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
		TaskQueue: "CART_TASK_QUEUE",
	}

	state := app.CartState{Items: make([]app.CartItem, 0)}
	we, err := c.ExecuteWorkflow(context.Background(), options, app.CartWorkflow, state)
	if err != nil {
		log.Fatalln("unable to execute workflow", err)
	}

	update := app.AddToCartSignal{Route: app.RouteTypes.ADD_TO_CART, Item: app.CartItem{ProductId:0, Quantity: 1}}
	err = c.SignalWorkflow(context.Background(), workflowID, "", "ADD_TO_CART_CHANNEL", update)

	resp, err := c.QueryWorkflow(context.Background(), workflowID, "", "getCart")
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
// @@@SNIPEND
