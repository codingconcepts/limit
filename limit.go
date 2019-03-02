package limit

import (
	"math"
	"strconv"
	"time"

	"github.com/go-redis/redis"
)

type limiter struct {
	client *redis.Client
}

type rate struct {
	client        *redis.Client
	windowSeconds float64
	windowNanos   int64
	maxCalls      int
}

// New returns an instance of the Redis wrapper.  It has no ability
// to rate-limit by itself and requires a call to Rate, which provides
// a specific sliding window period or a call to one of the HTTP
// handler implementations which themselves create and use an instance
// of the sliding window period.
func New(c *redis.Client) *limiter {
	return &limiter{
		client: c,
	}
}

// Rate returns a specific rate-limited period and provides the method
// that will determine whether an action is permitted or not, based on
// the rate limit in place.
func (l *limiter) Rate(maxCalls int, d time.Duration) *rate {
	return &rate{
		client:        l.client,
		windowSeconds: d.Seconds(),
		windowNanos:   d.Nanoseconds(),
		maxCalls:      maxCalls,
	}
}

// Allow determines whether an action is possible based on previous
// calls to Allow against the configured sliding window and maximum
// allowed calls.
func (r *rate) Allowed(id string) (bool, error) {
	now := time.Now().UnixNano()

	tx := r.client.TxPipeline()
	tx.ZRemRangeByScore(id, "0", strconv.FormatInt(now-r.windowNanos, 10))

	rangeCmd := tx.ZRange(id, 0, -1)
	tx.ZAdd(id, redis.Z{Score: float64(now), Member: now})
	tx.Expire(id, time.Duration(r.windowNanos)*time.Nanosecond)

	_, err := tx.Exec()
	if err != nil {
		return false, err
	}

	timestamps, err := rangeCmd.Result()
	if err != nil {
		return false, err
	}

	return math.Max(0, float64(r.maxCalls-len(timestamps))) > 0, nil
}
