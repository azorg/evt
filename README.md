# evt - simple in-process on-memory event bus based on Go channels

## Features
 - Object oriented lightweight API (only two class: `Bus` and `Sub`)
 - One default event broker (bus) out of the box
 - Two publish methods (directly and to inbox channel)
 - Flush method (including version with timeout)
 - Gracefull shutdown

## Block diagram of two event ways

### Short direct way (may blocking)
`Publish[Ex](topic, message)` -> Each subscriber channel -> Each subscriber

### Long way via inbox channel (non-blocking)
`PublishInbox[Ex](topic, message)` -> Common inbox channel -> Bus monitor ->
Each subscriber channel -> Each subscriber

## Examples

### Create new event bus (broker)
```go
  bus := evt.Bus(context.Background(), inboxChannelSize) // *evt.Bus
```

### Subsctibe to event topic
```go
  sub := bus.Subscribe("topic", channelSize) // *evt.Sub
  ...
  topic := sub.Topic() // get subscriber topic
  ...
  subscribed := sub.Subscribed() // check subscription (bool)
```

### Wait event
```go
  msg, ok := sub.Wait() // any, bool
  if !ok { // subscriber unsibscribed or bus canceled
   ...
  }
```

### Read from subscriber channel
```go
  select {
  case msg, ok := <-sub.C(): // any, bool
    if !ok { // subsciber unsibscribed or bus canceled
      ...
    }
    ...
  } // select
```

### Publish event to topic directly (may blocking)
```go
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
```

### Publish event to topic via inbox channel (buffered, non-blocking)
```go
  msg := "hello"
  bus.PublishInbox("topic", msg)
  ...
  msg = "world"
  bus.PublishInboxEx("topic", msg, 5*time.Second) // timeout = 5s
```

### Wait until all published message delivered (flush)
```go
  bus.Flush()
  ...
  err := bus.FlushEx(10*time.Secind) // timeout = 10s
  if err != nil {
    if errors.Is(err, evt.ErrTimeout) { // timeout
      ...
    }
  }
```

### Unsubscribe from event topic
```go
  sub.Cancel()
```

### Gracefull shutdown
```go
  // Cancel bus, unsubscribe all subscribers, cancel goroutines
  bus.Cancel()
	
  err := bus.WaitEx(10*time.Second) // timeout = 10s
  if err != nil {
    if errors.Is(err, evt.ErrTimeout) { // timeout
      ...
    }
  }
```

