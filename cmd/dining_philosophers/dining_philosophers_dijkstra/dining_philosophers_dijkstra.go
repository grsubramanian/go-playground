package main

import (
	"fmt"
	"github.com/grsubramanian/go-playground/internal"
	"math/rand"
	"sync"
	"time"
)

type DijkstraFork struct {
	sync.Mutex
}

type DijkstraPhilosopher struct {
	id         int
	firstFork  *DijkstraFork
	secondFork *DijkstraFork
}

func (philosopher DijkstraPhilosopher) dine() {
	philosopher.inState("thinking")
	internal.RandomPause(2)

	philosopher.inState("hungry")
	philosopher.firstFork.Lock()
	philosopher.secondFork.Lock()

	philosopher.inState("eating")
	internal.RandomPause(5)

	philosopher.secondFork.Unlock()
	philosopher.firstFork.Unlock()

	philosopher.dine()
}

func (philosopher DijkstraPhilosopher) inState(state string) {
	fmt.Printf("#%d is %s\n", philosopher.id, state)
}

func init() {
	// Random seed
	rand.Seed(time.Now().UTC().UnixNano())
}

func main() {
	count := 5

	forks := make([]*DijkstraFork, count)
	for i := 0; i < count; i++ {
		forks[i] = new(DijkstraFork)
	}

	philosophers := make([]*DijkstraPhilosopher, count)
	for i := 0; i < count; i++ {

		var leftFork, rightFork = forks[i], forks[(i+1)%count]

		var fistFork, secondFork *DijkstraFork
		if i == 0 {
			// For the first philosopher alone, we'll treat the right fork as the fork to pick up.
			fistFork, secondFork = rightFork, leftFork
		} else {
			// For most of the philosophers, we'll treat the left fork as the first fork to pick up.
			fistFork, secondFork = leftFork, rightFork
		}
		philosophers[i] =
			&DijkstraPhilosopher{
				id:         i,
				firstFork:  fistFork,
				secondFork: secondFork,
			}
		go philosophers[i].dine()
	}

	// Wait endlessly while they're dining
	endless := make(chan int)
	<-endless
}
