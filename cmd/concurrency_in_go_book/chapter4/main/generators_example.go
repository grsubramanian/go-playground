package main

import (
	"fmt"
	"github.com/gokul2411s/concurrency/pkg/concurrency_in_go_book/chapter4"
)

func main() {

	done := make(chan interface{})
	defer close(done)

	foreverOne := chapter4.Repeat(done, 1)
	only5Ones := chapter4.Take(done, foreverOne, 5)
	for i := range only5Ones {
		fmt.Printf("Val = %d\n", i)
	}

}
