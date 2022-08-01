package main

import (
	"flag"
	"fmt"
	"sync"
	"time"
)

type ChannelBasedButton struct {
	Clicked chan interface{}
}

func subscribeToChannelBasedButton(c chan interface{}, fn func()) {
	var goroutineRunning sync.WaitGroup
	goroutineRunning.Add(1)
	go func() {
		goroutineRunning.Done()

		// We add a sleep here to simulate a time gap between when the goroutineRunning wait group
		// is done and when we start blocking on the channel.
		time.Sleep(1*time.Second)

		<-c
		fn()
	}()
	goroutineRunning.Wait()
}

func main() {
	nWaitersFlag := flag.Int("n", 10, "Number of waiters")
	flag.Parse()

	button := ChannelBasedButton{
		Clicked: make(chan interface{}),
	}

	var clickRegistered sync.WaitGroup
	clickRegistered.Add(*nWaitersFlag)
	for i := 0; i < *nWaitersFlag; i++ {
		j := i
		subscribeToChannelBasedButton(
			button.Clicked,
			func() {
				fmt.Printf("Button click registered by %d\n", j)
				clickRegistered.Done()
			})
	}

	// Once a channel is closed, any reader from the channel will get a sentinel message.
	// Before the channel is closed, any reader will block, since there is no writer into the channel.
	close(button.Clicked)

	clickRegistered.Wait()
}
