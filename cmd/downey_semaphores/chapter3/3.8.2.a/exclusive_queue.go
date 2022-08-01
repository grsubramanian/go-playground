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

var previousLeaderDone, _ = internal.NewSemaphore(1, 1)
var previousFollowerDone, _ = internal.NewSemaphore(1, 1)

var leaderAvailable, _ = internal.NewSemaphore(1, 0)
var followerAvailable, _ = internal.NewSemaphore(1, 0)

func dance(id int, typ string) {
	time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)
	fmt.Printf("%s %d dancing\n", typ, id)
}

// Allows only one leader to be on stage at a given point in time.
func leaderCore(id int) {
	previousLeaderDone.Wait()

	leaderAvailable.Signal()
	followerAvailable.Wait()

	dance(id, "leader")

	previousLeaderDone.Signal()
}

// Allows only one follower to be on stage at a given point in time.
func followerCore(id int) {
	previousFollowerDone.Wait()

	followerAvailable.Signal()
	leaderAvailable.Wait()

	dance(id, "follower")

	previousFollowerDone.Signal()
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
