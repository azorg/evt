// File: "default.go"

package evt

import (
	"context"
	"time"
)

// Default inbox channel size
const DEFAULT_INBOX_SIZE = 1000

// Default event bus (one instance)
var bus *Bus

func init() {
	bus = New(context.Background(), DEFAULT_INBOX_SIZE)
}

// Get default event bus
func DefaultBus() *Bus {
	return bus
}

// Subsctibe to event topic
func Subscribe(topic string, size int) *Sub {
	return bus.Subscribe(topic, size)
}

// Get all subscribed topics
func Topics() []string {
	return bus.Topics()
}

// Get count of event subscribers
func Count(topic string) int {
	return bus.Count(topic)
}

// Get inbox channel
func C() chan<- Evt {
	return bus.C()
}

// Publish event to topic immediately (may blocking)
func Publish(topic string, msg any) (int, error) {
	return bus.Publish(topic, msg)
}

// Publish event to topic immediately with timeout (may blocking)
func PublishEx(topic string, msg any, timeout time.Duration) (int, error) {
	return bus.PublishEx(topic, msg, timeout)
}

// Publish event to topic via inbox channel (non-blocking)
func PublishInbox(topic string, msg any) {
	bus.Publish(topic, msg)
}

// Publish event to topic via inbox channel with timeout (non-blocking)
func PublishInboxEx(topic string, msg any, timeout time.Duration) {
	bus.PublishEx(topic, msg, timeout)
}

// Wait until all published message delivered
func Flush() {
	bus.Flush()
}

// Wait until all published message delivered with timeout
func FlushEx(timeout time.Duration) error {
	return bus.FlushEx(timeout)
}

// Cancel bus, unsubscribe all subscribers, cancel goroutines
func Cancel() {
	bus.Cancel()
}

// Wait until graceful shutdown
func Wait() {
	bus.Wait()
}

// Wait until graceful shutdown with timeout
func WaitEx(timeout time.Duration) error {
	return bus.WaitEx(timeout)
}

// EOF: "default.go"
