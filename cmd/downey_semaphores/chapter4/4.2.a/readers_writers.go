package main

import (
	"fmt"
	"sync"

	"github.com/grsubramanian/go-playground/internal"
)

var wg sync.WaitGroup

var r = 10
var w = 10
var c = 10

var readers = 0
var mutex, _ = internal.NewSemaphore(1, 1)

var roomEmpty, _ = internal.NewSemaphore(1, 1)

func enterReadCriticalSection() {
	mutex.Wait()
	readers++
	if readers == 1 {
		roomEmpty.Wait()
	}
	mutex.Signal()
}

func exitReadCriticalSection() {
	mutex.Wait()
	readers--
	if readers == 0 {
		roomEmpty.Signal()
	}
	mutex.Signal()
}

func read(id int, val int) {

	enterReadCriticalSection()

	fmt.Printf("Reader %d reading %d\n", id, val)

	exitReadCriticalSection()
}

func enterWriteCriticalSection() {
	roomEmpty.Wait()
}

func exitWriteCriticalSection() {
	roomEmpty.Signal()
}

func write(id int, val int) {

	enterWriteCriticalSection()

	fmt.Printf("Writer %d writing %d\n", id, val)

	exitWriteCriticalSection()
}

func reader(id int) {

	for i := 0; i < c; i++ {
		read(id, i)
	}

	wg.Done()
}

func writer(id int) {

	for i := 0; i < c; i++ {
		write(id, i)
	}

	wg.Done()
}

func main() {

	wg.Add(r + w)
	defer wg.Wait()

	for i := 0; i < w; i++ {
		go writer(i)
	}

	for i := 0; i < r; i++ {
		go reader(i)
	}
}
