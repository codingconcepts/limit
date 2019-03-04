package limit

import (
	"log"
	"os"
	"testing"

	"github.com/benbjohnson/clock"

	"github.com/alicebob/miniredis"
	"github.com/go-redis/redis"
)

var (
	mock      *miniredis.Miniredis
	client    *redis.Client
	limit     *limiter
	mockClock *clock.Mock
)

func TestMain(m *testing.M) {
	var err error
	if mock, err = miniredis.Run(); err != nil {
		log.Fatalf("error starting miniredis: %v", err)
	}

	client = redis.NewClient(&redis.Options{
		Addr:     mock.Addr(),
		Password: "",
		DB:       0,
	})

	if err := client.Ping().Err(); err != nil {
		log.Fatalf("error pinging redis: %v", err)
	}

	limit = New(client)
	mockClock = clock.NewMock()
	limit.clock = mockClock

	os.Exit(m.Run())
}
