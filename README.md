# limit
A sliding window rate-limited backed by Redis inspired by https://engagor.github.io/blog/2017/05/02/sliding-window-rate-limiter-redis/.

## Installation

```
$ go get -u github.com/codingconcepts/limit
```

## How it works

This library makes use of sorted sets in Redis for rate-limiting.  A call to `rate.Allowed` takes the ID of the unique caller you're limiting (e.g. an IP address) which will become its own sorted set in Redis.

As an example, we'll assume a unique ID of "8.8.8.8" and a rate that allows for 10 calls per second.  Every call to `Allowed` will perform the following in Redis:

* Start a transaction.
* Remove all items from the set that were performed over 1 second ago.
* Return all of the items in the sorted set (this allows us to determine if the limit has been reached).
* Add the current timestamp to the sorted set.
* Add an expiry of 1 second to the sorted set.
* Commit the transaction.

## Usage

### Basic

``` go
package main

import (
	"fmt"
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
	
	fmt.Println(rate.Allowed("8.8.8.8")) // true <nil>
	fmt.Println(rate.Allowed("8.8.8.8")) // true <nil>
	fmt.Println(rate.Allowed("8.8.8.8")) // true <nil>
	fmt.Println(rate.Allowed("8.8.8.8")) // false <nil>
}
```

### HTTP

``` go
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
```

## Todos

* Need a fallback mechanism, in case Redis can't be reached.