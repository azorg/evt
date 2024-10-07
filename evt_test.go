// File: "evt_test.go"

package evt

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

// Test default (global) broker
func TestDefault(t *testing.T) {
	var sub SubInterface = Subscribe("topic", 0)

	go func() {
		count, _ := Publish("topic", "hello")
		if count != 1 {
			t.Fatalf(`Publish("hello") return %d, want 1`, count)
		}

		count, _ = PublishEx("topic", "world", time.Second)
		if count != 1 {
			t.Fatalf(`Publish("world") return %d, want 1`, count)
		}
	}()

	msg, ok := <-sub.C()
	if !ok {
		t.Fatalf("msg, ok := <-sub.c(): return ok=false, want true")
	}

	str := msg.(string)
	if str != "hello" {
		t.Fatalf(`msg, ok := <-sub.c(): return "%s", want "hello"`, str)
	}
	fmt.Println("receive #1:", str)

	msg, ok = sub.Wait()
	if !ok {
		t.Fatalf("msg, ok := <-sub.c(): return ok=false, want true")
	}

	str = msg.(string)
	if str != "world" {
		t.Fatalf(`msg, ok := <-sub.c(): return "%s", want "world"`, str)
	}
	fmt.Println("receive #2:", str)

	topics := Topics()
	if num := len(topics); num != 1 {
		t.Fatalf("len(topics) return %d, want 1", num)
	}

	if topic := topics[0]; topic != "topic" {
		t.Fatalf(`topics[0] = "%s", wgant "topic"`, topic)
	}

	// Cancel default event broker
	start := time.Now()
	Cancel()

	err := WaitEx(time.Second)
	dt := time.Now().Sub(start)
	if err != nil {
		t.Fatalf("Wait(time.Second) return err=%v, want nil", err)
	}
	fmt.Printf("cancel: dt=%v\n", dt)

	topics = Topics()
	if num := len(topics); num != 0 {
		t.Fatalf("len(topics) return %d, want 0", num)
	}

	msg, ok = sub.Wait()
	if ok {
		t.Fatalf("sub.Wait(): return ok=true, want false")
	}
}

var wgSub sync.WaitGroup
var wgPub sync.WaitGroup

func goSub(sub *Sub, mod int, t *testing.T) {
	defer wgSub.Done()

	for {
		msg, ok := sub.Wait()
		if !ok {
			fmt.Println(sub.Topic(), "closed")
			return
		}

		i := msg.(int)
		if (i & 1) != mod {
			t.Fatalf("subscriber %s read %d", sub.Topic(), i)
		}
		fmt.Println(sub.Topic(), "<", i)
	} // for
}

func goPub(bus *Bus, topic string, start, num int, t *testing.T) {
	defer wgPub.Done()

	for i := start; i < num; i += 2 {
		start := time.Now()
		bus.PublishEx(topic, i, 1*time.Second) // FIXME: magic
		dt := time.Now().Sub(start)
		fmt.Println("publish:", topic, ">", i, "dt:", dt)
	} // for
}

// Test concurrency
func TestEvt(t *testing.T) {
	bus := New(context.Background(), 0)

	sub0 := bus.Subscribe("even", 0)
	sub1 := bus.Subscribe("odd", 0)
	sub2 := bus.Subscribe("even", 1)
	sub3 := bus.Subscribe("odd", 2)

	fmt.Println("topics:", bus.topics)
	for _, topic := range []string{"even", "odd"} {
		count := bus.Count(topic)
		if count != 2 {
			t.Fatalf("bus.Count(%s) return %d, want 2", topic, count)
		}
		fmt.Printf("bus.Count(%s)=%d\n", topic, count)
	}

	num := 20
	wgPub.Add(2)
	go goPub(bus, "even", 0, num, t)
	go goPub(bus, "odd", 1, num, t)

	wgSub.Add(4)
	go goSub(sub0, 0, t)
	go goSub(sub1, 1, t)
	go goSub(sub2, 0, t)
	go goSub(sub3, 1, t)

	wgPub.Wait()

	start := time.Now()
	bus.Cancel()
	err := bus.WaitEx(time.Second)
	dt := time.Now().Sub(start)
	if err != nil {
		t.Fatalf("bus.Wait(time.Second) return err=%v, want nil", err)
	}
	fmt.Printf("cancel: dt=%v\n", dt)

	wgSub.Wait()
}

// Test publish with timeout
func TestPublishEx(t *testing.T) {
	bus := New(context.Background(), 0)

	_ = bus.Subscribe("lazy", 0)

	start := time.Now()
	bus.PublishEx("lazy", "nothing", 1000*time.Millisecond)
	dt := time.Now().Sub(start)
	if dt < 900*time.Millisecond ||
		dt > 1100*time.Millisecond {
		t.Fatalf("bad delay: dt=%v, want 1s", dt)
	}
	fmt.Println("dt (1s):", dt)

	_ = bus.Subscribe("stupid", 0)
	start = time.Now()
	bus.PublishEx("stupid", "nothing", 500*time.Millisecond)
	bus.Flush()
	dt = time.Now().Sub(start)
	if dt < 400*time.Millisecond ||
		dt > 600*time.Millisecond {
		t.Fatalf("bad delay: dt=%v, want 500ms", dt)
	}
	fmt.Println("dt (500ms):", dt)
}

// EOF: "evt_test.go"
