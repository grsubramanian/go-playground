package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/grsubramanian/go-playground/internal"
)

var wg sync.WaitGroup

var a1Done, _ = internal.NewSemaphore(1, 0)
var b1Done, _ = internal.NewSemaphore(1, 0)

func doWork(val string) {
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	fmt.Println(val)
}

func signaller(done internal.Semaphore, val string) {
	doWork(val)

	// signal.
	done.Signal()
}

func waiter(done internal.Semaphore, val string) {

	// wait.
	done.Wait()

	doWork(val)
}

func a1() {
	signaller(a1Done, "a1")
}

func a2() {
	waiter(b1Done, "a2")
}

func threadA() {

	a1()
	a2()

	wg.Done()
}

func b1() {
	signaller(b1Done, "b1")
}

func b2() {
	waiter(a1Done, "b2")
}

func threadB() {

	b1()
	b2()

	wg.Done()
}

func main() {

	rand.Seed(time.Now().UnixNano())

	wg.Add(2)
	defer wg.Wait()

	go threadA()
	go threadB()
}
