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

<img src="https://codebarbarian-images.s3.amazonaws.com/shopping-cart.jpg">

With a traditional web app, abandoned cart notifications are tricky.
You need to use a job queue like [Celery](https://en.wikipedia.org/wiki/Celery_(software\)) in Python or [Machinery](https://github.com/RichardKnop/machinery) in Golang.
You need to schedule a job that checks if the cart is abandoned, and reschedule that job every time the cart is updated.

With Temporal, you don't need a separate job queue. Instead, you define a _Selector_ with two event handlers: one that responds to a Workflow signal, and one that responds to a timer.

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
		if !sentAbandonedCartEmail {
			selector.AddFuture(workflow.NewTimer(ctx, abandonedCartTimeout), func(f workflow.Future) {
				sentAbandonedCartEmail = true
				ao := workflow.ActivityOptions{
					ScheduleToStartTimeout: time.Minute,
					StartToCloseTimeout:    time.Minute,
				}

				ctx = workflow.WithActivityOptions(ctx, ao)

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