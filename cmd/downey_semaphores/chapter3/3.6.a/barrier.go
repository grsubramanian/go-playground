package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/grsubramanian/go-playground/internal"
)

var wg sync.WaitGroup

var n = 100
var count = 0
var countLock = sync.Mutex{}

var done, _ = internal.NewSemaphore(n, 0)

func rendezvous() {
	countLock.Lock()
	count++
	if count == n {
		done.SignalN(n)
	}
	countLock.Unlock()

	done.Wait()
}

func doWork(i int, val string) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	fmt.Printf("Thread %d %s\n", i, val)
}

func pre(i int) {
	doWork(i, "rendezvous")
}

func post(i int) {
	doWork(i, "criticalPoint")
}

func thread(i int) {
	defer wg.Done()

	pre(i)
	rendezvous()
	post(i)
}

func main() {

	rand.Seed(time.Now().UnixNano())

	wg.Add(n)
	for i := 0; i < n; i++ {
		go thread(i)
	}

	wg.Wait()
}
