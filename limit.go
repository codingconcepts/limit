package limit

import (
	"strconv"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/go-redis/redis"
)

type limiter struct {
	client redis.Cmdable
	clock  clock.Clock
}

type rate struct {
	client        redis.Cmdable
	clock         clock.Clock
	windowSeconds float64
	windowNanos   int64
	maxCalls      int
}

// New returns an instance of the Redis wrapper.  It has no ability
// to rate-limit by itself and requires a call to Rate, which provides
// a specific sliding window period or a call to one of the HTTP
// handler implementations which themselves create and use an instance
// of the sliding window period.
func New(c redis.Cmdable) *limiter {
	return &limiter{
		client: c,
		clock:  clock.New(),
	}
}

// Rate returns a specific rate-limited period and provides the method
// that will determine whether an action is permitted or not, based on
// the rate limit in place.
func (l *limiter) Rate(maxCalls int, d time.Duration) *rate {
	return &rate{
		client:        l.client,
		clock:         l.clock,
		windowSeconds: d.Seconds(),
		windowNanos:   d.Nanoseconds(),
		maxCalls:      maxCalls,
	}
}

// Allow determines whether an action is possible based on previous
// calls to Allow against the configured sliding window and maximum
// allowed calls.
func (r *rate) Allowed(id string) (allowed bool, left int, err error) {
	now := r.clock.Now().UnixNano()

	tx := r.client.TxPipeline()

	remCmd := tx.ZRemRangeByScore(id, "0", strconv.FormatInt(now-r.windowNanos, 10))
	if err = remCmd.Err(); err != nil {
		return false, 0, err
	}

	rangeCmd := tx.ZRange(id, 0, -1)
	if err = rangeCmd.Err(); err != nil {
		return false, 0, err
	}

	addCmd := tx.ZAdd(id, redis.Z{Score: float64(now), Member: now})
	if err = addCmd.Err(); err != nil {
		return false, 0, err
	}

	expCmd := tx.Expire(id, time.Duration(r.windowNanos)*time.Nanosecond)
	if err = expCmd.Err(); err != nil {
		return false, 0, err
	}

	if _, err = tx.Exec(); err != nil {
		return false, 0, err
	}

	timestamps, err := rangeCmd.Result()
	if err != nil {
		return false, 0, err
	}

	left = max(0, r.maxCalls-len(timestamps))
	return left > 0, left, nil
}

func max(x, y int) int {
	if x > y {
		return x
	}

	return y
}
