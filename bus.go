// File: "bus.go"

package evt

import (
	"context"
	"sync"
	"time"
)

// Subscibers set
type subs map[*Sub]struct{}

// Event
type Evt struct {
	Topic   string        // topic name
	Msg     any           // payload
	Timeout time.Duration // delivery timeout or zero
}

// Event bus (broker)
type Bus struct {
	topics map[string]subs // subscribers set for eatch topic
	mx     sync.RWMutex    // mutex for topics
	inbox  chan Evt        // inbox buffer channel
	ctx    context.Context // cancel context
	cancel func()          // context handler
	wgMon  sync.WaitGroup  // wait monitor goroutine
	wgPub  sync.WaitGroup  // wait publisher goroutines
	wgBuf  sync.WaitGroup  // wait inbox buffered message
}

// Event bus interface
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
	Publish(topic string, msg any) int

	// Publish event to topic immediately with timeout (may blocking)
	PublishEx(string, msg any) int

	// Publish event to topic via inbox channel (non-blocking)
	PublishInbox(topic string, msg any) bool

	// Publish event to topic via inbox channel with timeout (non-blocking)
	PublishInboxEx(topic string, msg any) bool

	// Wait until all published message delivered
	Flush()

	// Wait until all published message delivered with timeout
	FlushEx(timeout time.Duration) error

	// Cancel bus, unsubscribe all subscribers, cancel goroutines
	Cancel()

	// Wait until graceful shutdown
	Wait()

	// Wait until graceful shutdown with timeout
	WaitEx(timeout time.Duration) error
}

// Create new event bus (broker)
//
//	ctx - cancel context
//	inboxSize - inbox channel size
func New(ctx context.Context, inboxSize int) *Bus {
	bus := &Bus{
		topics: make(map[string]subs),
		inbox:  make(chan Evt, inboxSize),
	}
	bus.ctx, bus.cancel = context.WithCancel(ctx)
	bus.wgMon.Add(1)
	go bus.goMonitor()
	return bus
}

// Subsctibe to event topic
//
//	topic - event topic
//	size - channel size of subscribers
func (bus *Bus) Subscribe(topic string, size int) *Sub {
	sub := &Sub{
		topic: topic,
		bus:   bus,
		ch:    make(chan any, size),
	}

	bus.mx.Lock()
	defer bus.mx.Unlock()

	ss, ok := bus.topics[topic]
	if !ok { // no any subscribers
		ss = subs{}
		bus.topics[topic] = ss
	}

	ss[sub] = struct{}{} // add subscriber to set
	sub.pss = &ss
	return sub
}

// Get all subscribed topics
func (bus *Bus) Topics() []string {
	topics := make([]string, 0, len(bus.topics))

	bus.mx.RLock()
	defer bus.mx.RUnlock()

	for topic := range bus.topics {
		topics = append(topics, topic)
	}
	return topics
}

// Get count of event subscribers
func (bus *Bus) Count(topic string) int {
	bus.mx.RLock()
	defer bus.mx.RUnlock()

	subs, ok := bus.topics[topic]
	if !ok {
		return 0 // topic not found
	}
	return len(subs)
}

// Get inbox channel
func (bus *Bus) C() chan<- Evt {
	return bus.inbox
}

// Publish event to topic immediately (may blocking)
//
//	topic - event topic
//	msg - message (event payload)
//
//	count - actual number of topic subscribers
//	err - nil or context.Canceled
func (bus *Bus) Publish(topic string, msg any) (
	count int, err error) {

	bus.mx.RLock()
	defer bus.mx.RUnlock()

	ss, ok := bus.topics[topic]
	if !ok {
		return 0, nil // topic not found
	}

	// Send event to all topic subscribers
	for sub := range ss {
		select {
		case sub.ch <- msg: // write to subscriber channel
		case <-bus.ctx.Done(): // cancel by context
			return 0, context.Canceled
		} // select
		count++
	} // for

	return count, nil
}

// Publish event to topic immediately with timeout (may blocking)
//
//	 topic - event topic
//	 msg - message (event payload)
//		timeout - timeoit of write to each subscriber channel
//
//	 count - actual number of topic subscribers
//	 err - nil or context.Canceled or ErrTimeout
func (bus *Bus) PublishEx(
	topic string, msg any, timeout time.Duration) (
	count int, err error) {

	bus.mx.RLock()
	defer bus.mx.RUnlock()

	ss, ok := bus.topics[topic]
	if !ok {
		return 0, nil // topic not found
	}

	// Send event to all topic subscribers
	for sub := range ss {
		select {
		case sub.ch <- msg: // write to subscriber channel
		case <-bus.ctx.Done(): // cancel by context
			return 0, context.Canceled
		case <-time.After(timeout): // cancel by timeout
			return count, ErrTimeout
		} // select
		count++
	} // for

	return count, nil
}

// Publish event to topic via inbox channel (non-blocking)
//
//	topic - event topic
//	msg - message (event payload)
func (bus *Bus) PublishInbox(topic string, msg any) {
	bus.wgPub.Add(1)
	go func() {
		defer bus.wgPub.Done()
		select {
		case bus.inbox <- Evt{Topic: topic, Msg: msg}:
			bus.wgBuf.Add(1)
		case <-bus.ctx.Done(): // cancel by context
		} // select
	}()
}

// Publish event to topic via inbox channel with timeout (non-blocking)
//
//	topic - event topic
//	msg - message (event payload)
//	timeout - timeoit of write to each subscriber channel
func (bus *Bus) PublishInboxEx(topic string, msg any, timeout time.Duration) {
	bus.wgPub.Add(1)
	go func() {
		defer bus.wgPub.Done()
		select {
		case bus.inbox <- Evt{Topic: topic, Msg: msg, Timeout: timeout}:
			bus.wgBuf.Add(1)
		case <-bus.ctx.Done(): // cancel by context
		case <-time.After(timeout): // break by timeout
		} // select
	}()
}

// Wait until all published message delivered
func (bus *Bus) Flush() {
	bus.wgPub.Wait()
	bus.wgBuf.Wait()
}

// Wait until all published message delivered with timeout
func (bus *Bus) FlushEx(timeout time.Duration) error {
	if timeout == time.Duration(0) { // infinite wait
		bus.Flush()
		return nil
	}

	done := make(chan struct{})
	go func() {
		bus.Flush()
		close(done)
	}()

	select {
	case <-done:
		return nil // shutdown success
	case <-time.After(timeout):
		return ErrTimeout // timeout
	}
}

// Cancel bus, unsubscribe all subscribers, cancel goroutines
func (bus *Bus) Cancel() {
	bus.cancel()
}

// Wait until graceful shutdown
func (bus *Bus) Wait() {
	bus.wgPub.Wait()
	bus.wgMon.Wait()
}

// Wait until graceful shutdown with timeout
func (bus *Bus) WaitEx(timeout time.Duration) error {
	if timeout == time.Duration(0) { // infinite wait
		bus.Wait()
		return nil
	}

	done := make(chan struct{})
	go func() {
		bus.Wait()
		close(done)
	}()

	select {
	case <-done:
		return nil // shutdown success
	case <-time.After(timeout):
		return ErrTimeout // timeout
	}
}

// Enent bus monitor
func (bus *Bus) goMonitor() {
	defer bus.wgMon.Done()
	defer bus.shutdown()
	for {
		select {
		case <-bus.ctx.Done(): // cancel by context
			return

		case evt, ok := <-bus.inbox: // send event to subscribers
			if !ok { // its looks like cancel
				return
			}
			if evt.Timeout == time.Duration(0) {
				bus.Publish(evt.Topic, evt.Msg)
			} else {
				bus.PublishEx(evt.Topic, evt.Msg, evt.Timeout)
			}
			bus.wgBuf.Done()
		} // select
	} // for
}

// Shutdown bus: wait publisher goroutines, unsubscribe all subscribers,
// close inbox channel
func (bus *Bus) shutdown() {
	bus.wgPub.Wait() // wait publisher goroutines

	bus.mx.Lock()
	defer bus.mx.Unlock()

	// Unsubscribe all subscribers
	for topic, ss := range bus.topics {
		for sub := range ss {
			delete(ss, sub) // delete subscriber from set
			close(sub.ch)   // close subscriber channel

			// Mark as unsubscribed
			sub.bus = nil
			sub.pss = nil
		}
		delete(bus.topics, topic)
	} // for

	close(bus.inbox) // close inbox channel
}

// EOF: "bus.go"
