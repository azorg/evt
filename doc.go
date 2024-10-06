/*
evt - simple in-process on-memory event bus based on Go channels

# Create bus

	bus := evt.Bus()

# Subsctibe to event

	sub := bus.Sub("topic", 0) // 0 - channel size

# Wait event

	msg, ok := sub.Wait()
	if !ok { // subsciber unsibscribed of bus canceled
	  ...
	}

# Read from event channel

	select {
	case msg, ok := <-sub.C():
		if !ok { // subsciber unsibscribed or bus canceled
			...
		}
	...
	} // select

# Publish event to topic

	msg := "hello"
	bus.Pub("topic", msg)

# Ubsubscribe

	sub.Unsub()

# Cancel bus, unsubscribe all subscribers

	bus.Cancel()
*/
package evt

// EOF: "doc.go"
