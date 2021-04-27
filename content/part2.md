# Building an eCommerce Web App With Temporal, Part 2: Reminder Emails

In [Part 1](https://gist.github.com/vkarpov15/d0b4d3b1eb8ced160bd68172323eb379#file-part_1-md), you built out a simple shopping cart app using a long-lived Workflow to track the state of the cart.
Instead of storing the cart in a database, Temporal lets you represent the cart as a function invocation, using Signals to update the cart and Queries to get the state of the cart.

```golang
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

The long-lived Workflow pattern doesn't offer any substantial benefits in this simple CRUD app.
But there are certain tasks that are trivial with the long-lived Workflow pattern that would be much more difficult with a traditional web application.
For example, let's take a look at implementing an abandoned cart email notification.

Checking for an Abandoned Cart
------------------------------

In eCommerce, an [_abandoned_ shopping cart](https://www.optimizely.com/optimization-glossary/shopping-cart-abandonment/#:~:text=Shopping%20cart%20abandonment%20is%20when,process%20before%20completing%20the%20purchase.&text=This%20rate%20will%20identify%20what,don't%20complete%20the%20purchase.) is a shopping cart that has items, but the user hasn't added
any new items or checked out after a few hours.
Below is an example abandoned cart email that I recently received with an offer to incentivize checkout.

<img src="https://codebarbarian-images.s3.amazonaws.com/shopping-cart.jpg">

With a traditional web app, abandoned cart notifications are tricky.
You need to use a job queue like [Celery](https://en.wikipedia.org/wiki/Celery_(software\)) in Python or [Machinery](https://github.com/RichardKnop/machinery) in Golang.
You need to schedule a job that checks if the cart is abandoned, and reschedule that job every time the cart is updated.

With Temporal, you don't need a separate job queue. Instead, you define a _Selector_ with two event handlers: one that responds to a Workflow signal, and one that responds to a timer.
By creating a new Selector on each iteration of the `for` loop, you're telling Temporal to handle the next update cart signal it receives, or send an abandoned cart email if it doesn't receive a signal for `abandonedCartTimeout`.
Calling `Select()` on a selector blocks the Workflow until there's either a signal or `abandonedCartTimeout` elapses.

```golang
func CartWorkflow(ctx workflow.Context, state CartState) error {
	logger := workflow.GetLogger(ctx)

	err := workflow.SetQueryHandler(ctx, "getCart", func(input []byte) (CartState, error) {
		return state, nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed.", "Error", err)
		return err
	}

	channel := workflow.GetSignalChannel(ctx, "cartMessages"
	sentAbandonedCartEmail := false

	for {
    // Create a new selector on each iteration of the loop means Temporal will pick the first
    // event that occurs each time: either receiving a signal, or responding to the timer.
		selector := workflow.NewSelector(ctx)
		selector.AddReceive(channel, func(c workflow.ReceiveChannel, _ bool) {
			var signal interface{}
			c.Receive(ctx, &signal)

			// Handle signals for updating the cart
		})

    // If the user doesn't update the cart for `abandonedCartTimeout`, send an email
    // reminding them about their cart. Only send the email once.
		if !sentAbandonedCartEmail && len(state.Items) > 0 {
			selector.AddFuture(workflow.NewTimer(ctx, abandonedCartTimeout), func(f workflow.Future) {
				sentAbandonedCartEmail = true
				ao := workflow.ActivityOptions{
					StartToCloseTimeout:   10 * time.Second,
				}

				ctx = workflow.WithActivityOptions(ctx, ao)

        // More on SendAbandonedCartEmail in the next section
				err := workflow.ExecuteActivity(ctx, SendAbandonedCartEmail, state.Email).Get(ctx, nil)
				if err != nil {
					logger.Error("Error sending email %v", err)
					return
				}
			})
		}

		selector.Select(ctx)
	}

	return nil
}
```

Temporal Selectors make implementing an abandoned cart email trivial.
No need to implement a job queue, write a separate worker, or handle rescheduling jobs.
Just create a new Selector after every signal and use `AddFuture()` to handle the case where the user abandons their cart.
Temporal does the hard work of persisting and distributing the state of your Workflow for you.

Next up, let's take a closer look at Activities and the `ExecuteActivity()` call above that is responsible for sending the abandoned cart email.

Sending Emails from an Activity
-------------------------------

You can think of Activities as an abstraction for side effects in Temporal.
[Workflows should be pure, idempotent functions](https://docs.temporal.io/docs/go-create-workflows/#implementation) to allow Temporal to re-run a Workflow to recreate the Workflow's state.
Any side effects, like HTTP requests to the [Mailgun API](https://thecodebarbarian.com/sending-emails-using-the-mailgun-api.html), should be in an Activity.

For example, below is the implementation of the `SendAbandonedCartEmail()` function.
It loads Mailgun keys from environment variables, and sends an HTTP request to the Mailgun API using [Mailgun's official Go library](https://github.com/mailgun/mailgun-go).
The function takes two parameters: the workflow context, and the email as a string.

```golang
import (
	"context"
	"fmt"
	"github.com/mailgun/mailgun-go"
)

var (
	mailgunDomain = os.Getenv("MAILGUN_DOMAIN")
	mailgunKey    = os.Getenv("MAILGUN_PRIVATE_KEY")
)

func SendAbandonedCartEmail(_ context.Context, email string) error {
	mg := mailgun.NewMailgun(mailgunDomain, mailgunKey)
	m := mg.NewMessage(
		"noreply@"+mailgunDomain, // Sender
		"You've abandoned your shopping cart!", // Subject
		"Go to http://localhost:8080 to finish checking out!", // Placeholder email copy
		email, // Recipient
	)
	_, _, err := mg.Send(m)
	if err != nil {
		fmt.Println("Mailgun err: " + err.Error())
    return err
	}

	return err
}
```

As a reminder, below is the `ExecuteActivity()` call from the cart Workflow.
The 3rd parameter to `ExecuteActivity()` becomes the 2nd parameter to `SendAbandonedCartEmail()`:

```golang
workflow.ExecuteActivity(ctx, SendAbandonedCartEmail, state.Email).Get(ctx, nil)
```

The `ExecuteActivity()` function also exposes some neat options.
For example, [Temporal automatically retries failed activities](https://docs.temporal.io/docs/go-retries/), so Temporal would automatically retry the `SendAbandonedCart()` Activity up to 5 times if `SendAbandonedCart()` returns an error.
You can configure the number of times Temporal will retry the Activity in case of an error using the `MaximumAttempts` option as shown below.

```golang
ao := workflow.ActivityOptions{
	ScheduleToStartTimeout: time.Minute,
	StartToCloseTimeout:    time.Minute,
  MaximumAttempts:        3,
}

ctx = workflow.WithActivityOptions(ctx, ao)

err := workflow.ExecuteActivity(ctx, SendAbandonedCartEmail, state.Email).Get(ctx, nil)
```

Moving On
---------

Long-lived Workflows in Temporal are excellent for scheduled tasks.
You can build durable time-based logic, like checking whether the user hasn't modified their shopping cart for a given period of time, without using a job queue.
Next up, we'll look at patterns for build RESTful APIs on top of Temporal Workflows.
