package internal

import (
	"math/rand"
	"time"
)

func RandomPause(max int) {
	time.Sleep(time.Millisecond * time.Duration(rand.Intn(max*1000)))
}
