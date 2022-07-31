package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var wg sync.WaitGroup

var a1Done sync.WaitGroup
var b1Done sync.WaitGroup

func doWork(val string) {
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	fmt.Println(val)
}

func signaller(done *sync.WaitGroup, val string) {
	doWork(val)

	// signal.
	done.Done()
}

func waiter(done *sync.WaitGroup, val string) {

	// wait.
	done.Wait()

	doWork(val)
}

func a1() {
	signaller(&a1Done, "a1")
}

func a2() {
	waiter(&b1Done, "a2")
}

func threadA() {

	a1()
	a2()

	wg.Done()
}

func b1() {
	signaller(&b1Done, "b1")
}

func b2() {
	waiter(&a1Done, "b2")
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

	a1Done.Add(1)
	b1Done.Add(1)

	go threadA()
	go threadB()
}
