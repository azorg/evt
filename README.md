# evt - simple in-process on-memory event bus based on Go channels

## Features
 - Object oriented lightweight API (only two class: `Bus` and `Sub`)
 - One default event broker (bus) out of the box (look `DefaultBus()`)
 - Non blocking publish
 - Gracefull shutdown

## Examples

### Create new event bus (broker)
```go
  bus := evt.Bus(context.Background()) // *evt.Bus
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
  if !ok { // subscriber unsibscribed of bus canceled
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

### Publish event to topic (non-blocking)
```go
  msg := "hello"
  bus.Publish("topic", msg)
```

### Unsubscribe from event topic
```go
  sub.Cancel()
```

### Gracefull shutdown
```go
  // Cancel bus, unsubscribe all subscribers, cancel publisher goroutines
  bus.Cancel()
	
  ok := bus.Wait(time.Second) // timeout = 1s
  if !ok { // timeout
    ...
  }
```

