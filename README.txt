package evt // import "github.com/azorg/evt"

evt - simple in-process on-memory event bus based on Go channels

# Features

  - Object oriented lightweight API (only two class: `Bus` and `Sub`)
  - One default event broker (bus) out of the box
  - Two publish methods (directly and buffered to inbox channel)
  - Flush method (including version with timeout)
  - Graceful shutdown

# Block diagram of two event ways

 1. Short direct way (may blocking) `Publish[Ex](topic, message)` -> Each
    subscriber channel -> Each subscriber

 2. Long buffered way via inbox channel (non-blocking) `PublishInbox[Ex](topic,
    message)` -> Common inbox channel -> Bus monitor -> Each subscriber channel
    -> Each subscriber

# Examples

    // Create new event bus (broker)
    // ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
    bus := evt.Bus(context.Background(), inboxChannelSize) // *evt.Bus

    // Subscribe to event topic
    sub := bus.Subscribe("topic", channelSize) // *evt.Sub
    ...
    topic := sub.Topic() // get subscriber topic
    ...
    subscribed := sub.Subscribed() // check subscription (bool)

    // Wait event
    // ^^^^^^^^^^
    msg, ok := sub.Wait() // any, bool
    if !ok { // subscriber unsubscribed or bus canceled
    ...
    }

    // Read from subscriber channel
    // ^^^^^^^^^^^^^^^^^^^^^^^^^^^^
    select {
    case msg, ok := <-sub.C(): // any, bool
    	if !ok { // subscriber unsubscribed or bus canceled
    	...
    	}
    ...
    } // select

    // Publish event to topic directly (may blocking)
    // ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
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
    // ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
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
    // ^^^^^^^^^^^^^^^^^^^^^^^^^^^^
    sub.Cancel()

    // Graceful shutdown
    // ^^^^^^^^^^^^^^^^^
    // Cancel bus, unsubscribe all subscribers, cancel goroutines
    bus.Cancel()

    // Wait until graceful shutdown with timeout (wait goroutines finished)
    err := bus.WaitEx(10*time.Second) // timeout = 10s
    if err != nil {
    	if errors.Is(err, evt.ErrTimeout) { // timeout
    		...
    	}
    }

CONSTANTS

const DefaultInboxSize = 1000
    Default inbox channel size


VARIABLES

var ErrTimeout = errors.New("timeout")

FUNCTIONS

func C() chan<- Evt
    Get inbox channel

func Cancel()
    Cancel bus, unsubscribe all subscribers, cancel goroutines

func Count(topic string) int
    Get count of event subscribers

func Flush()
    Wait until all published message delivered

func FlushEx(timeout time.Duration) error
    Wait until all published message delivered with timeout

func Publish(topic string, msg any) (int, error)
    Publish event to topic immediately (may blocking)

func PublishEx(topic string, msg any, timeout time.Duration) (int, error)
    Publish event to topic immediately with timeout (may blocking)

func PublishInbox(topic string, msg any)
    Publish event to topic via inbox channel (non-blocking)

func PublishInboxEx(topic string, msg any, timeout time.Duration)
    Publish event to topic via inbox channel with timeout (non-blocking)

func Topics() []string
    Get all subscribed topics

func Wait()
    Wait until graceful shutdown (wait goroutines finished)

func WaitEx(timeout time.Duration) error
    Wait until graceful shutdown with timeout (wait goroutines finished)


TYPES

type Bus struct {
	// Has unexported fields.
}
    Event bus (broker)

func DefaultBus() *Bus
    Get default event bus

func New(ctx context.Context, inboxSize int) *Bus
    Create new event bus (broker)

        ctx - cancel context
        inboxSize - inbox channel size

func (bus *Bus) C() chan<- Evt
    Get inbox channel

func (bus *Bus) Cancel()
    Cancel bus, unsubscribe all subscribers, cancel goroutines

func (bus *Bus) Count(topic string) int
    Get count of event subscribers

func (bus *Bus) Flush()
    Wait until all published message delivered

func (bus *Bus) FlushEx(timeout time.Duration) error
    Wait until all published message delivered with timeout

func (bus *Bus) Publish(topic string, msg any) (
	count int, err error)
    Publish event to topic immediately (may blocking)

        topic - event topic
        msg - message (event payload)

        count - actual number of topic subscribers
        err - nil or context.Canceled

func (bus *Bus) PublishEx(
	topic string, msg any, timeout time.Duration) (
	count int, err error)
    Publish event to topic immediately with timeout (may blocking)

        topic - event topic
        msg - message (event payload)
        timeout - timeoit of write to each subscriber channel

        count - actual number of topic subscribers
        err - nil or context.Canceled or ErrTimeout

func (bus *Bus) PublishInbox(topic string, msg any)
    Publish event to topic via inbox channel (non-blocking)

        topic - event topic
        msg - message (event payload)

func (bus *Bus) PublishInboxEx(topic string, msg any, timeout time.Duration)
    Publish event to topic via inbox channel with timeout (non-blocking)

        topic - event topic
        msg - message (event payload)
        timeout - timeoit of write to each subscriber channel

func (bus *Bus) Subscribe(topic string, size int) *Sub
    Subsctibe to event topic

        topic - event topic
        size - channel size of subscriber

func (bus *Bus) Topics() []string
    Get all subscribed topics

func (bus *Bus) Wait()
    Wait until graceful shutdown (wait goroutines finished)

func (bus *Bus) WaitEx(timeout time.Duration) error
    Wait until graceful shutdown with timeout (wait goroutines finished)

type BusInterface interface {
	// Subscribe to topic event
	Subscribe(topic string, size int) *Sub

	// Get all subscribed topics
	Topics() []string

	// Get count of event subscribers
	Count(topic string) int

	// Get inbox channel
	C() chan<- Evt

	// Publish event to topic immediately (may blocking)
	Publish(topic string, msg any) (
		count int, err error)

	// Publish event to topic immediately with timeout (may blocking)
	PublishEx(topic string, msg any, timeout time.Duration) (
		count int, err error)

	// Publish event to topic via inbox channel (non-blocking)
	PublishInbox(topic string, msg any)

	// Publish event to topic via inbox channel with timeout (non-blocking)
	PublishInboxEx(topic string, msg any, timeout time.Duration)

	// Wait until all published message delivered
	Flush()

	// Wait until all published message delivered with timeout
	FlushEx(timeout time.Duration) error

	// Cancel bus, unsubscribe all subscribers, cancel goroutines
	Cancel()

	// Wait until graceful shutdown (wait goroutines finished)
	Wait()

	// Wait until graceful shutdown with timeout (wait goroutines finished)
	WaitEx(timeout time.Duration) error
}
    Event bus interface

type Evt struct {
	Topic   string        // topic name
	Msg     any           // payload
	Timeout time.Duration // delivery timeout or zero
}
    Event

type Sub struct {
	// Has unexported fields.
}
    Subscriber handler

func Subscribe(topic string, size int) *Sub
    Subscribe to event topic

func (sub *Sub) C() <-chan any
    Get subscriber channel

func (sub *Sub) Cancel()
    Unsubscribe from event topic

func (sub *Sub) Subscribed() bool
    Check subscription

func (sub *Sub) Topic() string
    Return subscriber topic

func (sub *Sub) Wait() (msg any, ok bool)
    Wait event (read from channel)

type SubInterface interface {
	Topic() string            // return subscriber topic
	Subscribed() bool         // check subscription
	C() <-chan any            // get subscriber channel
	Wait() (msg any, ok bool) // wait event (read from channel)
	Cancel()                  // unsubscribe from topic
}
    Subscriber interface

