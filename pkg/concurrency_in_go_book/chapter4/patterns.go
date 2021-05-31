package chapter4

import "sync"

var FanIn = func(done <-chan interface{}, channels ...<-chan interface{}) <-chan interface{} {

	out := make(chan interface{})

	var wg sync.WaitGroup
	wg.Add(len(channels))

	multiplex := func(c <-chan interface{}) {
		defer wg.Done()
		for i := range c {
			select {
			case <-done:
				return
			case out<- i:
			}
		}
	}

	for _, c := range channels {
		go multiplex(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

var OrDone = func(done, c <-chan interface{}) <-chan interface{} {

	out := make(chan interface{})
	go func() {
		defer close(out)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				select {
				case <-done:
				case out<- v:
				}
			}
		}
	}()
	return out
}

var Tee = func(done, c <-chan interface{}) (<-chan interface{}, <-chan interface{}) {

	out1 := make(chan interface{})
	out2 := make(chan interface{})

	go func() {
		defer close(out1)
		defer close(out2)

		for v := range OrDone(done, c) {
			var out1Cpy, out2Cpy = out1, out2
			for i := 0; i < 2; i++ {
				select {
				case <-done:
					return
				case out1Cpy<- v:
					out1Cpy = nil
				case out2Cpy<- v:
					out2Cpy = nil
				}
			}
		}
	}()

	return out1, out2
}

var Bridge = func(done <-chan interface{}, chanStream <-chan <-chan interface{}) <-chan interface{} {
	out := make(chan interface{})

	go func() {
		defer close(out)

		for {
			var valStream <-chan interface{}
			select {
			case <-done:
				return
			case v, ok := <-chanStream:
				if !ok {
					return
				}
				valStream = v
			}
			for v := range OrDone(done, valStream) {
				select {
				case <-done:
					return
				case out<- v:
				}
			}
		}
	}()

	return out
}