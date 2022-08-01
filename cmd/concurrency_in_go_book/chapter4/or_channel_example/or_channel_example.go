package main

import (
	"fmt"
	"github.com/grsubramanian/go-playground/pkg/concurrency_in_go_book/chapter4"
	"time"
)

func main() {
	timeout := func(after time.Duration) <-chan interface{} {
		out := make(chan interface{})
		go func() {
			defer close(out)
			time.Sleep(after)
		}()
		return out
	}

	startTime := time.Now()
	quickestTimeout := chapter4.Or(
		timeout(1*time.Second),
		timeout(1*time.Minute),
		timeout(1*time.Hour))

	<-quickestTimeout
	fmt.Printf("Time for completion: %v\n", time.Since(startTime))
}
