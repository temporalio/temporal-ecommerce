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
	}
)

func CartWorkflow(ctx workflow.Context, state CartState) error {
	logger := workflow.GetLogger(ctx)

	err := workflow.SetQueryHandler(ctx, "getCart", func(input []byte) (CartState, error) {
		return state, nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed.", "Error", err)
		return err
	}

	// TripCh to wait on trip completed event signals
	channel := workflow.GetSignalChannel(ctx, "addToCart")
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(channel, func(c workflow.ReceiveChannel, _ bool) {
		var toAdd CartItem
		c.Receive(ctx, &toAdd)
		state.Items = append(state.Items, toAdd)
	})

	for {
		selector.Select(ctx)
	}

	return nil
}
