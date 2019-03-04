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
	fmt.Println(rate.Allowed("8.8.8.8")) // true 3 <nil>
	fmt.Println(rate.Allowed("8.8.8.8")) // true 2 <nil>
	fmt.Println(rate.Allowed("8.8.8.8")) // true 1 <nil>
	fmt.Println(rate.Allowed("8.8.8.8")) // false 0 <nil>
}
