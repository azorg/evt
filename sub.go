// File: "sub.go"

package evt

// Subsciber handler
type Sub struct {
	topic string   // subscriber topic
	bus   *Bus     // pointer to parent bus
	pss   *subs    // pointer to subscribers set of this topic
	ch    chan any // event channel
}

// Subsctiber interface
type SubInterface interface {
	Topic() string            // return subscriber topic
	Subscribed() bool         // check subscription
	C() <-chan any            // get subscriber channel
	Wait() (msg any, ok bool) // wait event (read from channel)
	Cancel()                  // unsubsctibe from topic
}

// Return subscriber topic
func (sub *Sub) Topic() string {
	return sub.topic
}

// Check subscription
func (sub *Sub) Subscribed() bool {
	bus.mx.RLock()
	defer bus.mx.RUnlock()
	return sub != nil && sub.bus != nil && sub.pss != nil
}

// Get subscriber channel
func (sub *Sub) C() <-chan any {
	return sub.ch
}

// Wait event (read from channel)
func (sub *Sub) Wait() (msg any, ok bool) {
	msg, ok = <-sub.ch
	return // msg, ok
}

// Unsubscribe from event topic
func (sub *Sub) Cancel() {
	sub.bus.mx.Lock()
	defer sub.bus.mx.Unlock()

	if sub == nil || sub.bus == nil || sub.pss == nil {
		return // already unsibscribed
	}

	delete(*sub.pss, sub) // delete subscriber from set
	close(sub.ch)         // close event channel

	// Mark as unsubscribed
	sub.bus = nil
	sub.pss = nil
}

// EOF: "sub.go"
