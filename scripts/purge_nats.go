package main

import (
	"context"
	"fmt"
	"log"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func main() {
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()

	js, err := jetstream.New(nc)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	streams := []string{"WEBENCODE_WORK", "WEBENCODE_EVENTS"}
	for _, s := range streams {
		str, err := js.Stream(ctx, s)
		if err == nil {
			err = str.Purge(ctx)
			if err == nil {
				fmt.Printf("Purged stream %s\n", s)
			} else {
				fmt.Printf("Failed to purge %s: %v\n", s, err)
			}
		} else {
			fmt.Printf("Stream %s not found: %v\n", s, err)
		}
	}
}
