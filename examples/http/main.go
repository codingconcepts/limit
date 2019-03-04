package main

import (
	"log"
	"net/http"
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

	limiter := limit.New(client)

	// With a rate of 1 every 10 seconds.
	http.Handle("/one", limiter.LimitFuncHandler(2, time.Second*10, one))

	// With a rate of 5 every second.
	http.Handle("/two", limiter.LimitFuncHandler(5, time.Second, two))
	http.ListenAndServe(":1234", nil)
}

func one(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("one"))
}

func two(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("two"))
}
