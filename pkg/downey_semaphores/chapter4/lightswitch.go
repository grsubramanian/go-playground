package chapter4

import "github.com/grsubramanian/go-playground/internal"

type LightSwitch interface {
	Lock(s internal.Semaphore)
	Unlock(s internal.Semaphore)
}

type lightSwitchImpl struct {
	count int
	mutex internal.Semaphore
}

func NewLightSwitch() LightSwitch {
	m, _ := internal.NewSemaphore(1, 1)
	return &lightSwitchImpl{
		count: 0,
		mutex: m,
	}
}

func (l *lightSwitchImpl) Lock(s internal.Semaphore) {
	l.mutex.Wait()
	l.count++
	if l.count == 1 {
		s.Wait()
	}
	l.mutex.Signal()
}

func (l *lightSwitchImpl) Unlock(s internal.Semaphore) {
	l.mutex.Wait()
	l.count--
	if l.count == 0 {
		s.Signal()
	}
	l.mutex.Signal()
}
