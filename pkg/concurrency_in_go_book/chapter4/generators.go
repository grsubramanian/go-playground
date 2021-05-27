package chapter4

/**
 * A pre-emptible read-only channel that streams the input integers in order.
 */
var IntStream = func(done <-chan interface{}, vals ...int) <-chan int {

	out := make(chan int)
	go func() {
		defer close(out)
		for _, i := range vals {
			select {
			case <-done:
				return
			case out<- i:
			}
		}
	}()
	return out
}

/**
 * A pre-emptible read-only channel that streams the input integer repeatedly forever.
 */
var Repeat = func(done <-chan interface{}, val int) <-chan int {

	out := make(chan int)
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case out<- val:
			}
		}
	}()
	return out
}

/**
 * A pr-emptible read-only channel that streams only the first 'n' integers from the input stream.
 */
var Take = func(done <-chan interface{}, in <-chan int, n int) <-chan int {

	out := make(chan int)
	go func() {
		defer close(out)
		for i := 0; i < n; i++ {
			select {
			case <-done:
				return
			case out<- <-in:
			}
		}
	}()
	return out
}
