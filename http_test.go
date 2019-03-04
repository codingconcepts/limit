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

	// Request at 1000000ms.
	mockClock.Add(time.Millisecond)
	req, resp = httptest.NewRequest(http.MethodGet, "/", nil), httptest.NewRecorder()
	handler.ServeHTTP(resp, req)
	test.Equals(t, http.StatusOK, resp.Code)

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
}
