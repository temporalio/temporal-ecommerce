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

	UpdateCartMessage struct {
		Remove bool
		Item   CartItem
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

	channel := workflow.GetSignalChannel(ctx, "updateCart")
	selector := workflow.NewSelector(ctx)

	selector.AddReceive(channel, func(c workflow.ReceiveChannel, _ bool) {
		var update UpdateCartMessage
		c.Receive(ctx, &update)
		if update.Remove {
			RemoveFromCart(&state, update.Item)
		} else {
			AddToCart(&state, update.Item)
		}
	})

	for {
		selector.Select(ctx)
	}

	return nil
}

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
