// File: "bus.go"

package evt

import (
	"context"
	"sync"
	"time"
)

// Subscibers set
type subs map[*Sub]struct{}

// Event bus (broker)
type Bus struct {
	topics map[string]subs // subscribers set for eatch topic
	mx     sync.RWMutex
	ctx    context.Context
	cancel func()
	wgPub  sync.WaitGroup
	wgWait sync.WaitGroup
}

// Event bus interface
type BusInterface interface {
	Subscribe(topic string, size int) *Sub // subscribe to topic event
	Publish(topic string, msg any) int     // publish event to topic (non-blocking)
	Topics() []string                      // get all subscribed topics
	Count(topic string) int                // get count of subscribers
	Cancel()                               // cancel bus, unsubscribe all subscribers
	Wait(timeout time.Duration) error      // wait until graceful shutdown
}

// Create new event bus (broker)
func New(ctx context.Context) *Bus {
	bus := &Bus{topics: make(map[string]subs)}
	bus.ctx, bus.cancel = context.WithCancel(ctx)
	bus.wgWait.Add(1)
	go bus.goWaitCancel()
	return bus
}

// Subsctibe to event topic
func (bus *Bus) Subscribe(topic string, size int) *Sub {
	sub := &Sub{
		topic: topic,
		bus:   bus,
		ch:    make(chan any, size),
	}

	bus.mx.Lock()
	defer bus.mx.Unlock()

	ss, ok := bus.topics[topic]
	if !ok { // no any subscibers
		ss = subs{}
		bus.topics[topic] = ss
	}

	ss[sub] = struct{}{} // add subsciber to set
	sub.pss = &ss
	return sub
}

// Publish event to topic (non-blocking)
//
//	topic - event topic
//	msg - message (event payload)
func (bus *Bus) Publish(topic string, msg any) {
	bus.wgPub.Add(1)
	go func() {
		defer bus.wgPub.Done()
		bus.mx.RLock()
		defer bus.mx.RUnlock()

		ss, ok := bus.topics[topic]
		if !ok {
			return // topic not found
		}

		// Send event to all subscibers
		for sub := range ss {
			select {
			case sub.ch <- msg: // write to subscriber channel
			case <-bus.ctx.Done(): // cancel by context
			}
		}
	}()
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

// Cancel bus, unsubscribe all subscribers, cancel publisher goroutines
func (bus *Bus) Cancel() {
	bus.cancel()
}

// Wait until graceful shutdown, return false on timeout
func (bus *Bus) Wait(timeout time.Duration) bool {
	done := make(chan struct{})
	go func() {
		bus.wgWait.Wait()
		close(done)
	}()

	select {
	case <-done:
		return true // shutdown success
	case <-time.After(timeout):
		return false // timeout
	}
}

// Wait context cancel
func (bus *Bus) goWaitCancel() {
	defer bus.wgWait.Done()
	<-bus.ctx.Done()
	bus.wgPub.Wait()

	bus.mx.Lock()
	defer bus.mx.Unlock()

	for topic, ss := range bus.topics {
		for sub := range ss {
			delete(ss, sub) // delete subscriber from set
			close(sub.ch)   // close event channel

			// Mark as unsubscribed
			sub.bus = nil
			sub.pss = nil
		}
		delete(bus.topics, topic)
	}
}

// EOF: "bus.go"
