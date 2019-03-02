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

	http.Handle("/", limiter.LimitFuncHandler(3, time.Second*10, handler))
	http.ListenAndServe(":1234", nil)
}

func handler(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("hello"))
}
