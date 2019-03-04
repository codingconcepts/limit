package main

import (
	"fmt"
	"log"
	"time"

	"github.com/codingconcepts/limit"
	"github.com/go-redis/redis"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	if err := client.Ping().Err(); err != nil {
		log.Fatalf("error pinging redis: %v", err)
	}

	// Create a limiter allowing 3 calls every second.
	rate := limit.New(client).Rate(3, time.Second)

	for {
		allowed, left, err := rate.Allowed("8.8.8.8")
		if err != nil {
			log.Println("error determining rate-limit")
		}
		log.Printf("left: %d allowed: %v\n", left, allowed)
		fmt.Scanln()
	}
}
