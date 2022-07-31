package internal

import "errors"

type empty struct{}
type Semaphore chan empty

/*
Creates a new POSIX-style semaphore.

maxSignallers - the maximum number of signallers that need to be supported.
If more than these many signallers signal without being awaited,
then the signallers will block.

initial - the initial value of the semaphore. Should not be any more than the maximum
number of signallers.
*/
func NewSemaphore(maxSignallers int, initial int) (Semaphore, error) {
	if maxSignallers < 0 {
		return nil, errors.New("maximum number of signallers should be non-negative")
	}

	if initial > maxSignallers {
		return nil, errors.New("initial value should be less than or equal to maximum number of signallers")
	}

	s := make(Semaphore, maxSignallers)

	for i := 0; i < initial; i++ {
		s.Signal()
	}
	return s, nil
}

func (s Semaphore) Wait() {
	<-s
}

func (s Semaphore) Signal() {
	s <- empty{}
}
