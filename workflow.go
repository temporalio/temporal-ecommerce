package app

import (
	"github.com/mitchellh/mapstructure"
	"go.temporal.io/sdk/workflow"
	"time"
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

	channel := workflow.GetSignalChannel(ctx, "cartMessages")
	checkedOut := false

	for {
		var signal interface{}
		_ = channel.Receive(ctx, &signal)

		var routeSignal RouteSignal
		err := mapstructure.Decode(signal, &routeSignal)
		if err != nil {
			logger.Error("Invalid signal type %v", err)
			continue
		}

		switch {
		case routeSignal.Route == RouteTypes.ADD_TO_CART:
			var message AddToCartSignal
			err := mapstructure.Decode(signal, &message)
			if err != nil {
				logger.Error("Invalid signal type %v", err)
				continue
			}

			AddToCart(&state, message.Item)
		case routeSignal.Route == RouteTypes.REMOVE_FROM_CART:
			var message RemoveFromCartSignal
			err := mapstructure.Decode(signal, &message)
			if err != nil {
				logger.Error("Invalid signal type %v", err)
				continue
			}

			RemoveFromCart(&state, message.Item)
		case routeSignal.Route == RouteTypes.CHECKOUT:
			var message CheckoutSignal
			err := mapstructure.Decode(signal, &message)
			if err != nil {
				logger.Error("Invalid signal type %v", err)
				continue
			}

			state.Email = message.Email

			ao := workflow.ActivityOptions{
				ScheduleToStartTimeout: time.Minute,
				StartToCloseTimeout:    time.Minute,
			}

			ctx = workflow.WithActivityOptions(ctx, ao)

			err = workflow.ExecuteActivity(ctx, CreateStripeCharge, state).Get(ctx, nil)
			if err != nil {
				logger.Error("Invalid signal type %v", err)
				continue
			}

			checkedOut = true
		}

		if checkedOut {
			break
		}
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
