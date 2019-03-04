package limit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/codingconcepts/limit/test"
)

func TestLimitFuncHandler(t *testing.T) {
	handler := limit.LimitFuncHandler(2, time.Second, func(w http.ResponseWriter, r *http.Request) {})

	// Request at 0ms.
	req, resp := httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	test.Equals(t, http.StatusOK, resp.Code)
	assertResponseHeaders(t, resp.Header(), req, "2", "2", "1s")

	// Request at 1000000ms.
	mockClock.Add(time.Millisecond)
	req, resp = httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	test.Equals(t, http.StatusOK, resp.Code)
	assertResponseHeaders(t, resp.Header(), req, "2", "1", "1s")

	// Request at 2000000ms is disallowed because this is the third
	// request within 2 seconds.
	mockClock.Add(time.Millisecond)
	req, resp = httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	test.Equals(t, http.StatusTooManyRequests, resp.Code)

	// Request at 2002000000ms is allowed because this is outside
	// the 2 second window and hence part of the next window.
	mockClock.Add(time.Second * 2)
	req, resp = httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	test.Equals(t, http.StatusOK, resp.Code)
	assertResponseHeaders(t, resp.Header(), req, "2", "2", "1s")
}

func assertResponseHeaders(t *testing.T, h http.Header, r *http.Request, tot, left, dur string) {
	test.Equals(t, tot, h.Get(headerRateLimitTotal))
	test.Equals(t, left, h.Get(headerRateLimitRemaining))
	test.Equals(t, dur, h.Get(headerRateLimitDuration))
	test.Equals(t, "", h.Get(headerRateLimitForwardedFor))
	test.Equals(t, r.RemoteAddr, h.Get(headerRateLimitRemoteAddr))
}
