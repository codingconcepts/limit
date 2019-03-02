package limit

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// Error allows rate-specific information to be returned to the caller
// in the event that the limit has been reached.
type Error struct {
	Code          int     `json:"-"`
	MaxCalls      int     `json:"maxCalls"`
	WindowSeconds float64 `json:"windowSeconds"`
}

// Error returns the JSON representation of the Error object or an empty
// string in the case of a marshalling error.
func (e Error) Error() string {
	j, err := json.MarshalIndent(e, "", "  ")
	if err != nil {
		return ""
	}

	return string(j)
}

// LimitFuncHandler adds rate-limiting to a handler.
func (l *limiter) LimitFuncHandler(maxCalls int, d time.Duration, next http.HandlerFunc) http.Handler {
	return l.LimitHandler(maxCalls, d, http.HandlerFunc(next))
}

// LimitHandler adds rate-limiting to a handler.
func (l *limiter) LimitHandler(maxCalls int, d time.Duration, next http.Handler) http.Handler {
	r := l.Rate(maxCalls, d)

	middle := func(w http.ResponseWriter, req *http.Request) {
		if err := r.limitByRequest(w, req); err != nil {
			if lerr, ok := err.(Error); ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(lerr.Code)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
			}

			w.Write([]byte(err.Error()))
			return
		}

		next.ServeHTTP(w, req)
	}

	return http.HandlerFunc(middle)
}

func (r *rate) limitByRequest(w http.ResponseWriter, req *http.Request) error {
	r.setResponseHeaders(w, req)

	id := fmt.Sprintf("%s:%s", req.RemoteAddr, req.URL.Path)
	ok, err := r.Allowed(id)
	if err != nil {
		return err
	}
	if !ok {
		return Error{
			Code:          http.StatusTooManyRequests,
			MaxCalls:      r.maxCalls,
			WindowSeconds: r.windowSeconds,
		}
	}

	return nil
}

func (r *rate) setResponseHeaders(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("X-Rate-Limit-Limit", strconv.Itoa(r.maxCalls))
	w.Header().Add("X-Rate-Limit-Duration", strconv.FormatFloat(r.windowSeconds, 'f', 0, 64))
	w.Header().Add("X-Rate-Limit-Request-Forwarded-For", req.Header.Get("X-Forwarded-For"))
	w.Header().Add("X-Rate-Limit-Request-Remote-Addr", req.RemoteAddr)
}
