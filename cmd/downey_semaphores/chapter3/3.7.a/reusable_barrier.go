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
var allDone = false
var countLock = sync.Mutex{}

var barrierDone, _ = internal.NewSemaphore(n, 0)
var barrierReset, _ = internal.NewSemaphore(n, 0)

func barrier() {
	countLock.Lock()
	count++
	if count == n {
		barrierDone.SignalN(n)
	}
	countLock.Unlock()

	barrierDone.Wait()
}

func resetBarrier(i int) {
	countLock.Lock()
	count--
	if count == 0 {
		barrierReset.SignalN(n)
	}
	countLock.Unlock()

	barrierReset.Wait()
}

func doWork(repeat int, id int, val string) {
	time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
	fmt.Printf("Repeat %d thread %d %s\n", repeat, id, val)
}

func pre(repeat int, id int) {
	doWork(repeat, id, "rendezvous")
}

func post(repeat int, id int) {
	doWork(repeat, id, "criticalPoint")
}

func thread(id int, repeats int) {
	defer wg.Done()

	for repeat := 0; repeat < repeats; repeat++ {
		pre(repeat, id)
		barrier()
		post(repeat, id)
		resetBarrier(id)
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	wg.Add(n)
	for i := 0; i < n; i++ {
		go thread(i, 10)
	}

	wg.Wait()
}
