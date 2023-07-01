package main

import (
	"flag"
	"log"
	"math/rand"
	"time"

	"github.com/grsubramanian/go-playground/pkg/redblacktree"
	"github.com/pkg/errors"
)

var (
	seed     int64
	numItems int
)

func init() {
	flag.Int64Var(&seed, "seed", time.Now().Unix(), "seed for repeatable testing")
	flag.IntVar(&numItems, "N", 5, "number of items to insert into tree")
	flag.Parse()
}

func main() {

	log.Printf("Seed is %d", seed)

	r := rand.New(rand.NewSource(seed))
	input := r.Perm(numItems)

	t := &redblacktree.Tree[int]{}
	for _, val := range input {
		t.Insert(val)

		log.Printf("Tree after inserting %d", val)
		t.Print()

		err := t.CheckInvariants()
		if err != nil {
			log.Fatalln(errors.Wrapf(err, "validation failed"))
		}
	}
}
