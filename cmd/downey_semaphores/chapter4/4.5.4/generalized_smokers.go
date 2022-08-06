package main

import (
	"fmt"
	"sync"

	"github.com/grsubramanian/go-playground/internal"
)

var wg sync.WaitGroup

var s = 100

var primaryMaterial, _ = internal.NewSemaphore(2*s, 0)
var secondaryMaterial, _ = internal.NewSemaphore(2*s, 0)
var tertiaryMaterial, _ = internal.NewSemaphore(2*s, 0)

func materialsProducer(material1 internal.Semaphore, material2 internal.Semaphore, name string) {
	for i := 0; i < s; i++ {
		fmt.Printf("%s materials producer producing...\n", name)
		material1.Signal()
		material2.Signal()
	}
	wg.Done()
}

func nonPrimaryMaterialsProducer() {
	materialsProducer(secondaryMaterial, tertiaryMaterial, "Non primary")
}

func nonSecondaryMaterialsProducer() {
	materialsProducer(tertiaryMaterial, primaryMaterial, "Non secondary")
}

func nonTertiaryMaterialsProducer() {
	materialsProducer(primaryMaterial, secondaryMaterial, "Non tertiary")
}

var mutex, _ = internal.NewSemaphore(1, 1)
var numPrimaryMaterialInstancesStashed = 0
var numSecondaryMaterialInstancesStashed = 0
var numTertiaryMaterialInstancesStashed = 0

var nonPrimaryMaterialsAvailable, _ = internal.NewSemaphore(2*s, 0)
var nonSecondaryMaterialsAvailable, _ = internal.NewSemaphore(2*s, 0)
var nonTertiaryMaterialsAvailable, _ = internal.NewSemaphore(2*s, 0)

func pusher(
	material1Name string,
	material2Name string,
	material3Name string,
	numMaterial1InstancesStashed *int,
	numMaterial2InstancesStashed *int,
	numMaterial3InstancesStashed *int,
	material1 internal.Semaphore,
	materials1And2Available internal.Semaphore,
	materials1And3Available internal.Semaphore) {

	for true {
		material1.Wait()
		mutex.Wait()
		if (*numMaterial2InstancesStashed) > 0 {
			fmt.Printf("Consumer pusher found %s and %s materials\n", material1Name, material2Name)
			(*numMaterial2InstancesStashed)--
			materials1And2Available.Signal()
		} else if (*numMaterial3InstancesStashed) > 0 {
			fmt.Printf("Consumer pusher found %s and %s materials\n", material1Name, material3Name)
			(*numMaterial3InstancesStashed)--
			materials1And3Available.Signal()
		} else {
			fmt.Printf("Consumer pusher found only %s material\n", material1Name)
			(*numMaterial1InstancesStashed)++
		}
		mutex.Signal()
	}
}

func pusherForConsumerConsumingPrimaryMaterial() {
	pusher(
		"primary",
		"secondary",
		"tertiary",
		&numPrimaryMaterialInstancesStashed,
		&numSecondaryMaterialInstancesStashed,
		&numTertiaryMaterialInstancesStashed,
		primaryMaterial,
		nonTertiaryMaterialsAvailable,
		nonSecondaryMaterialsAvailable)
}

func pusherForConsumerConsumingSecondaryMaterial() {
	pusher(
		"secondary",
		"tertiary",
		"primary",
		&numSecondaryMaterialInstancesStashed,
		&numTertiaryMaterialInstancesStashed,
		&numPrimaryMaterialInstancesStashed,
		secondaryMaterial,
		nonPrimaryMaterialsAvailable,
		nonTertiaryMaterialsAvailable)
}

func pusherForConsumerConsumingTertiaryMaterial() {
	pusher(
		"tertiary",
		"primary",
		"secondary",
		&numTertiaryMaterialInstancesStashed,
		&numPrimaryMaterialInstancesStashed,
		&numSecondaryMaterialInstancesStashed,
		tertiaryMaterial,
		nonSecondaryMaterialsAvailable,
		nonPrimaryMaterialsAvailable)
}

func materialsConsumer(materialsAvailable internal.Semaphore, name string) {
	for i := 0; i < s; i++ {
		materialsAvailable.Wait()
		fmt.Printf("%s materials consumer consuming...\n", name)
	}
	wg.Done()
}

func nonPrimaryMaterialsConsumer() {
	materialsConsumer(nonPrimaryMaterialsAvailable, "Non primary")
}

func nonSecondaryMaterialsConsumer() {
	materialsConsumer(nonSecondaryMaterialsAvailable, "Non secondary")
}

func nonTertiaryMaterialsConsumer() {
	materialsConsumer(nonTertiaryMaterialsAvailable, "Non tertiary")
}

func main() {

	wg.Add(6)
	defer wg.Wait()

	// Producers.
	go nonPrimaryMaterialsProducer()
	go nonSecondaryMaterialsProducer()
	go nonTertiaryMaterialsProducer()

	// Consumers.
	go nonPrimaryMaterialsConsumer()
	go nonSecondaryMaterialsConsumer()
	go nonTertiaryMaterialsConsumer()

	// Pushers.
	go pusherForConsumerConsumingPrimaryMaterial()
	go pusherForConsumerConsumingSecondaryMaterial()
	go pusherForConsumerConsumingTertiaryMaterial()
}
