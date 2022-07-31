package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var wg sync.WaitGroup

var ctx context.Context = context.TODO()
var a1IsDone = false
var a1DoneCond = sync.NewCond(&sync.Mutex{})
var b1IsDone = false
var b1DoneCond = sync.NewCond(&sync.Mutex{})

func doWork(val string) {
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	fmt.Println(val)
}

func signaller(c *sync.Cond, cval *bool, val string) {
	doWork(val)

	// signal.
	c.L.Lock()
	*cval = true
	c.Signal()
	c.L.Unlock()
}

func waiter(c *sync.Cond, cval *bool, val string) {

	// wait.
	c.L.Lock()
	for !(*cval) {
		c.Wait()
	}
	c.Signal()
	c.L.Unlock()

	doWork(val)
}

func a1() {
	signaller(a1DoneCond, &a1IsDone, "a1")
}

func a2() {
	waiter(b1DoneCond, &b1IsDone, "a2")
}

func threadA() {

	a1()
	a2()

	wg.Done()
}

func b1() {
	signaller(b1DoneCond, &b1IsDone, "b1")
}

func b2() {
	waiter(a1DoneCond, &a1IsDone, "b2")
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
