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

var mutex, _ = internal.NewSemaphore(1, 1)

var room1 = 0
var room1Exit, _ = internal.NewSemaphore(1, 1)
var room2 = 0
var room2Exit, _ = internal.NewSemaphore(1, 0)

func lockLockWithNoStarvation() {
	// Entering room 1.
	mutex.Wait()
	room1++
	mutex.Signal()

	// Exiting room 1 and entering room 2.
	room1Exit.Wait()
	room2++

	mutex.Wait()
	room1--
	mutex.Signal()

	if room1 > 0 {
		room1Exit.Signal()
	} else {
		room2Exit.Signal()
	}

	// Exiting room 2.
	room2Exit.Wait()
	room2--

	// And now can do critical section.
}

func unlockLockWithNoStarvation() {
	if room2 > 0 {
		room2Exit.Signal()
	} else {
		room1Exit.Signal()
	}
}

func criticalSection(id int) {
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	fmt.Printf("Thread %d critical section\n", id)
}

func thread(id int) {
	lockLockWithNoStarvation()
	criticalSection(id)
	unlockLockWithNoStarvation()

	wg.Done()
}

func main() {

	wg.Add(n)
	defer wg.Wait()

	for i := 0; i < n; i++ {
		go thread(i)
	}
}
