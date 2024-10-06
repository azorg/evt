// File: "default.go"

package evt

import (
	"context"
	"time"
)

// Default event bus (one instance)
var bus *Bus

func init() {
	bus = New(context.Background())
}

// Get default event bus
func DefaultBus() *Bus {
	return bus
}

// Subsctibe to event topic
func Subscribe(topic string, size int) *Sub {
	return bus.Subscribe(topic, size)
}

// Get count of event subscribers
func Count(topic string) int {
	return bus.Count(topic)
}

// Publish event to topic (non-blocking),
// return number of actual subscribers
//
//	topic - event topic
//	msg - message (event payload)
func Publish(topic string, msg any) int {
	return bus.Publish(topic, msg)
}

// Get all subscribed topics
func Topics() []string {
	return bus.Topics()
}

// Cancel bus, unsubscribe all subscribers, cancel publisher goroutines
func Cancel() {
	bus.Cancel()
}

// Wait until graceful shutdown, return false on timeout
func Wait(timeout time.Duration) bool {
	return bus.Wait(timeout)
}

// EOF: "default.go"
