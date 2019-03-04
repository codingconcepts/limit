package limit

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	headerRateLimitTotal        = "X-Rate-Limit-Total"
	headerRateLimitRemaining    = "X-Rate-Limit-Remaining"
	headerRateLimitDuration     = "X-Rate-Limit-Duration"
	headerRateLimitForwardedFor = "X-Rate-Limit-Forwarded-For"
	headerRateLimitRemoteAddr   = "X-Rate-Limit-Remote-Addr"
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
	// Try to extract just the host name from the remote address, falling back
	// to the entire remote address if an error occurs.
	host, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		host = req.RemoteAddr
	}

	id := fmt.Sprintf("%s:%s", host, req.URL.Path)
	ok, left, err := r.Allowed(id)
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

	// Set the response header so the user knows how many requests they've got left.
	r.responseHeaders(w, req, left)

	return nil
}

func (r *rate) responseHeaders(w http.ResponseWriter, req *http.Request, left int) {
	w.Header().Add(headerRateLimitTotal, strconv.Itoa(r.maxCalls))
	w.Header().Add(headerRateLimitRemaining, strconv.Itoa(left))
	w.Header().Add(headerRateLimitDuration, fmt.Sprintf("%.0fs", r.windowSeconds))
	w.Header().Add(headerRateLimitForwardedFor, getIPAdress(req))
	w.Header().Add(headerRateLimitRemoteAddr, req.RemoteAddr)
}

func getIPAdress(r *http.Request) string {
	for _, h := range []string{"X-Forwarded-For", "X-Real-Ip"} {
		addresses := strings.Split(r.Header.Get(h), ",")

		// March from right to left until we get a public address that will
		// be the address right before our proxy.
		for i := len(addresses) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(addresses[i])

			// Headers can contain spaces, strip those out.
			realIP := net.ParseIP(ip)
			if !realIP.IsGlobalUnicast() {
				continue
			}
			return ip
		}
	}
	return ""
}
