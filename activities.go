package app

import (
	"context"
	"fmt"
	"github.com/mailgun/mailgun-go"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/charge"
)

type Activities struct {
	stripeKey string
	mailgunDomain string
	mailgunKey string
}

func MakeActivities(stripeKey string, mailgunDomain string, mailgunKey string) (*Activities) {
	return &Activities{
		stripeKey: stripeKey,
		mailgunDomain: mailgunDomain,
		mailgunKey: mailgunKey,
	}
}

func (a *Activities) CreateStripeCharge(_ context.Context, cart CartState) error {
	stripe.Key = a.stripeKey
	var amount float32 = 0
	var description string = ""
	for _, item := range cart.Items {
		var product Product
		for _, _product := range Products {
			if _product.Id == item.ProductId {
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

	_, err := charge.New(&stripe.ChargeParams{
		Amount:       stripe.Int64(int64(amount * 100)),
		Currency:     stripe.String(string(stripe.CurrencyUSD)),
		Description:  stripe.String(description),
		Source:       &stripe.SourceParams{Token: stripe.String("tok_visa")},
		ReceiptEmail: stripe.String(cart.Email),
	})

	if err != nil {
		fmt.Println("Stripe err: " + err.Error())
	}

	return err
}

func (a *Activities) SendAbandonedCartEmail(_ context.Context, email string) error {
	mg := mailgun.NewMailgun(a.mailgunDomain, a.mailgunKey)
	m := mg.NewMessage(
		"noreply@"+a.mailgunDomain,
		"You've abandoned your shopping cart!",
		"Go to http://localhost:8080 to finish checking out!",
		email,
	)
	_, _, err := mg.Send(m)
	if err != nil {
		fmt.Println("Mailgun err: " + err.Error())
		return err
	}

	return err
}
