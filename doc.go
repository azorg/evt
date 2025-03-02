// File: "doc.go"

/*
evt - simple in-process on-memory event bus based on Go channels

# Features

  - Object oriented lightweight API (only two class: `Bus` and `Sub`)
  - One default event broker (bus) out of the box
  - Two publish methods (directly and buffered to inbox channel)
  - Flush method (including version with timeout)
  - Graceful shutdown

# Block diagram of two event ways

 1. Short direct way (may blocking)
    `Publish[Ex](topic, message)` -> Each subscriber channel -> Each subscriber

 2. Long buffered way via inbox channel (non-blocking)
    `PublishInbox[Ex](topic, message)` -> Common inbox channel -> Bus monitor ->
    Each subscriber channel -> Each subscriber

# Examples

	  // Create new event bus (broker)
		// -----------------------------
		bus := evt.Bus(context.Background(), inboxChannelSize) // *evt.Bus

		// Subscribe to event topic
		sub := bus.Subscribe("topic", channelSize) // *evt.Sub
		...
		topic := sub.Topic() // get subscriber topic
		...
		subscribed := sub.Subscribed() // check subscription (bool)

		// Wait event
		// ----------
		msg, ok := sub.Wait() // any, bool
		if !ok { // subscriber unsubscribed or bus canceled
		...
		}

		// Read from subscriber channel
		// ----------------------------
		select {
		case msg, ok := <-sub.C(): // any, bool
			if !ok { // subscriber unsubscribed or bus canceled
			...
			}
		...
		} // select

		// Publish event to topic directly (may blocking)
		// ----------------------------------------------
		msg := "hello"
		count, err := bus.Publish("topic", msg)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				...
			}
		}
		if count == 0 { // no any subscribers, message lost
			...
		}
		...
		msg = "world"
		count, err = bus.PublishEx("topic", msg, 3*time.Second) // timeout = 3s
		if err != nil {
			if errors.Is(err, context.Canceled) {
				...
			} else if errors.Is(err, evt.ErrTimeout) {
				...
			}
		}

		// Publish event to topic via inbox channel (buffered, non-blocking)
		// -----------------------------------------------------------------
		msg := "hello"
		bus.PublishInbox("topic", msg)
		...
		msg = "world"
		bus.PublishInboxEx("topic", msg, 5*time.Second) // timeout = 5s

		// Wait until all published message delivered (flush)
		bus.Flush()
		...
		err := bus.FlushEx(10*time.Secind) // timeout = 10s
		if err != nil {
			if errors.Is(err, evt.ErrTimeout) { // timeout
				...
			}
		}

		// Unsubscribe from event topic
		// ----------------------------
		sub.Cancel()

		// Graceful shutdown
		// -----------------
		// Cancel bus, unsubscribe all subscribers, cancel goroutines
		bus.Cancel()

		// Wait until graceful shutdown with timeout (wait goroutines finished)
		err := bus.WaitEx(10*time.Second) // timeout = 10s
		if err != nil {
			if errors.Is(err, evt.ErrTimeout) { // timeout
				...
			}
		}
*/
package evt

// EOF: "doc.go"
