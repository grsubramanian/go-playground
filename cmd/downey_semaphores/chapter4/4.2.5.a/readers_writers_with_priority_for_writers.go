package main

import (
	"fmt"
	"sync"

	"github.com/grsubramanian/go-playground/internal"
	dsc "github.com/grsubramanian/go-playground/pkg/downey_semaphores/chapter4"
)

var wg sync.WaitGroup

var r = 10
var w = 10
var c = 10

var readPermitter = dsc.NewLightSwitch()
var writePermitter = dsc.NewLightSwitch()

var readPermission, _ = internal.NewSemaphore(1, 1)
var writePermission, _ = internal.NewSemaphore(1, 1)

func enterReadCriticalSection() {
	readPermission.Wait()
	readPermitter.Lock(writePermission)
	readPermission.Signal()
}

func exitReadCriticalSection() {
	readPermitter.Unlock(writePermission)
}

func read(id int, val int) {

	enterReadCriticalSection()

	fmt.Printf("Reader %d reading %d\n", id, val)

	exitReadCriticalSection()
}

func enterWriteCriticalSection() {
	writePermitter.Lock(readPermission)
	writePermission.Wait()
}

func exitWriteCriticalSection() {
	writePermission.Signal()
	writePermitter.Unlock(readPermission)
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

	for i := 0; i < r; i++ {
		go reader(i)
	}

	for i := 0; i < w; i++ {
		go writer(i)
	}
}
