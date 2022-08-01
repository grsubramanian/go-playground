package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/grsubramanian/go-playground/internal"
)

var wg sync.WaitGroup

// The number of followers, also the number of leaders.
// So, total 2*n folks.
var n = 10

var unpairedLeaders = 0
var unpairedFollowers = 0

var leaderAvailable, _ = internal.NewSemaphore(1, 0)
var followerAvailable, _ = internal.NewSemaphore(1, 0)

var rendezvous, _ = internal.NewSemaphore(1, 0)

var lock, _ = internal.NewSemaphore(1, 1)

func dance(id int, typ string) {
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	fmt.Printf("%s %d dancing\n", typ, id)
}

func leaderCore(id int) {

	lock.Wait()

	if unpairedFollowers > 0 {
		unpairedFollowers--
		leaderAvailable.Signal()
	} else {
		unpairedLeaders++
		lock.Signal()
		followerAvailable.Wait()
	}

	dance(id, "leader")

	rendezvous.Signal()

	lock.Signal()
}

func followerCore(id int) {

	lock.Wait()

	if unpairedLeaders > 0 {
		unpairedLeaders--
		followerAvailable.Signal()
	} else {
		unpairedFollowers++
		lock.Signal()
		leaderAvailable.Wait()
	}

	dance(id, "follower")

	rendezvous.Wait()
}

func leader(id int) {
	leaderCore(id)
	wg.Done()
}

func follower(id int) {
	followerCore(id)
	wg.Done()
}

func main() {

	rand.Seed(time.Now().UnixNano())

	wg.Add(2 * n)
	defer wg.Wait()

	for i := 0; i < n; i++ {
		go leader(i)
		go follower(i)
	}

}
