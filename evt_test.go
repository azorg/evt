// File: "evt_test.go"

package evt

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestDefault(t *testing.T) {
	var sub SubInterface = Subscribe("topic", 0)

	count := Publish("topic", "hello")
	if count != 1 {
		t.Fatalf(`Publish("hello") return %d, want 1`, count)
	}

	msg, ok := <-sub.C()
	if !ok {
		t.Fatalf("msg, ok := <-sub.c(): return ok=false, want true")
	}

	str := msg.(string)
	if str != "hello" {
		t.Fatalf(`msg, ok := <-sub.c(): return "%s", want "hello"`, str)
	}
	fmt.Println("receive:", str)

	topics := Topics()
	if num := len(topics); num != 1 {
		t.Fatalf("len(topics) return %d, want 1", num)
	}

	if topic := topics[0]; topic != "topic" {
		t.Fatalf(`topics[0] = "%s", wgant "topic"`, topic)
	}

	// Cancel dafaul event broker
	start := time.Now()
	Cancel()

	ok = Wait(time.Second)
	dt := time.Now().Sub(start)
	if !ok {
		t.Fatalf("Wait(time.Second) return %v, want true", ok)
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
		bus.Publish(topic, i)
		dt := time.Now().Sub(start)
		fmt.Println(topic, ">", i, "dt:", dt)
	} // for
}

func TestEvt(t *testing.T) {
	bus := New(context.Background())

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

	num := 10
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
	ok := bus.Wait(time.Second)
	dt := time.Now().Sub(start)
	if !ok {
		t.Fatalf("bus.Wait(time.Second) return %v, want true", ok)
	}
	fmt.Printf("cancel: dt=%v\n", dt)

	wgSub.Wait()
}

// EOF: "evt_test.go"
