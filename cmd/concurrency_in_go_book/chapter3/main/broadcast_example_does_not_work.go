package main

import (
    "flag"
    "fmt"
    "sync"
    "time"
)

type Button struct {
    Clicked *sync.Cond
}

func subscribe(c *sync.Cond, fn func()) {
    var goroutineRunning sync.WaitGroup
    goroutineRunning.Add(1)
    go func() {
        goroutineRunning.Done()
        time.Sleep(1*time.Second) // added to simulate the issue more regularly.
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

    button := Button{
        Clicked: sync.NewCond(&sync.Mutex{}),
    }

    var clickRegistered sync.WaitGroup
    clickRegistered.Add(*nWaitersFlag)
    for i := 0; i < *nWaitersFlag; i++ {
        j := i
        subscribe(
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
