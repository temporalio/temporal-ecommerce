package app

import (
	"context"
	"fmt"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/charge"
	// "go.temporal.io/sdk/activity"
	"os"
)

var (
	stripeKey = os.Getenv("STRIPE_PRIVATE_KEY")
	// temporal client.Client
)

func CreateStripeCharge(_ context.Context, cart CartState) error {
	stripe.Key = stripeKey
	var amount float32 = 0
	var description string = ""
	for _, item := range cart.Items {
		var product Product
		for _, _product := range Products {
			if (_product.Id == item.ProductId) {
				product = _product
				break
			}
		}
		amount += float32(item.Quantity) * product.Price
		if len(description) > 0 {
			description += ", "
		}
		description += product.Name
	}

	_, err := charge.New(&stripe.ChargeParams {
		Amount:       stripe.Int64(int64(amount * 100)),
		Currency:     stripe.String(string(stripe.CurrencyUSD)),
		Description:  stripe.String(description),
		Source:       &stripe.SourceParams{Token: stripe.String("tok_visa")},
		ReceiptEmail: stripe.String(cart.Email),
	})

	return err
}