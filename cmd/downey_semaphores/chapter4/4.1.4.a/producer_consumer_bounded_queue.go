package main

import (
	"fmt"
	"sync"

	"github.com/grsubramanian/go-playground/internal"
)

var wg sync.WaitGroup

var n = 10

var nonEmpty, _ = internal.NewSemaphore(n, 0)
var nonFull, _ = internal.NewSemaphore(n, n)

var lock = sync.RWMutex{}

var q = internal.NewQueue()

func producerCore(item int) {

	nonFull.Wait()

	lock.Lock()
	q.Enqueue(item)
	fmt.Printf("Enqueued item %d\n", item)
	lock.Unlock()

	nonEmpty.Signal()
}

func consumerCore() {

	nonEmpty.Wait()

	lock.Lock()
	item := q.Dequeue()
	fmt.Printf("Dequeued item %d\n", item)
	lock.Unlock()

	nonFull.Signal()
}

func producer() {
	for i := 0; i < 1000; i++ {
		producerCore(i)
	}

	wg.Done()
}

func consumer() {
	for i := 0; i < 1000; i++ {
		consumerCore()
	}

	wg.Done()
}

func main() {

	wg.Add(2)
	defer wg.Wait()

	go consumer()
	go producer()
}
