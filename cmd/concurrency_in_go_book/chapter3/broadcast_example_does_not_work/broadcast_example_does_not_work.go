package main

import (
    "flag"
    "fmt"
    "sync"
    "time"
)

type CVBasedButton struct {
    Clicked *sync.Cond
}

func subscribeToCVBasedButton(c *sync.Cond, fn func()) {
    var goroutineRunning sync.WaitGroup
    goroutineRunning.Add(1)
    go func() {
        goroutineRunning.Done()

        // We add a sleep here to simulate a time gap between when the goroutineRunning wait group
        // is done and when we start waiting on the condition variable.
        time.Sleep(1*time.Second)

        c.L.Lock()
        defer c.L.Unlock()
        c.Wait()
        fn()
    }()
    goroutineRunning.Wait()
}

func main() {
    nWaitersFlag := flag.Int("n", 10, "Number of waiters")
    flag.Parse()

    button := CVBasedButton{
        Clicked: sync.NewCond(&sync.Mutex{}),
    }

    var clickRegistered sync.WaitGroup
    clickRegistered.Add(*nWaitersFlag)
    for i := 0; i < *nWaitersFlag; i++ {
        j := i
        subscribeToCVBasedButton(
            button.Clicked,
            func() {
                fmt.Printf("Button click registered by %d\n", j)
                clickRegistered.Done()
            })
    }

    // This will wake up all waiters, however, not all subscribed handlers may be waiting.
    button.Clicked.Broadcast()

    clickRegistered.Wait()
}
