package tsclient

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

var fastRetryBackoffs = []time.Duration{1 * time.Millisecond, 1 * time.Millisecond, 1 * time.Millisecond}

func newTestClient(t *testing.T, serverURL string) *HTTPClient {
	t.Helper()
	c := New(
		WithTrustedAppKey("test-app", "test-key"),
		WithWharfEndpoint(serverURL),
	).(*HTTPClient)
	c.retryBackoffs = fastRetryBackoffs
	return c
}

func addTestMetric(t *testing.T, c *HTTPClient) {
	t.Helper()
	def := NewDefinition("test_metric")
	if err := c.AddMetric(def, 42.0); err != nil {
		t.Fatalf("AddMetric failed: %v", err)
	}
}

func TestFlush_ImmediateSuccess(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	addTestMetric(t, c)

	if err := c.Flush(); err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if n := atomic.LoadInt32(&attempts); n != 1 {
		t.Fatalf("expected 1 request (no retries), got %d", n)
	}
}

func TestFlush_EventualSuccess_After429(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n == 1 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	addTestMetric(t, c)

	if err := c.Flush(); err != nil {
		t.Fatalf("expected success after retry, got %v", err)
	}
	if n := atomic.LoadInt32(&attempts); n != 2 {
		t.Fatalf("expected 2 requests (1 initial + 1 retry), got %d", n)
	}
}

func TestFlush_MaxRetriesExhausted(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	addTestMetric(t, c)

	err := c.Flush()
	if err == nil {
		t.Fatal("expected error after exhausting retries, got nil")
	}
	httpErr, ok := err.(*UnexpectedHTTPStatusError)
	if !ok {
		t.Fatalf("expected *UnexpectedHTTPStatusError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected status 429, got %d", httpErr.StatusCode)
	}
	// 1 initial + 3 retries = 4 total attempts
	if n := atomic.LoadInt32(&attempts); n != 4 {
		t.Fatalf("expected 4 requests (1 initial + 3 retries), got %d", n)
	}
}

func TestFlush_NonRetryableError(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	addTestMetric(t, c)

	err := c.Flush()
	if err == nil {
		t.Fatal("expected error for 500, got nil")
	}
	httpErr, ok := err.(*UnexpectedHTTPStatusError)
	if !ok {
		t.Fatalf("expected *UnexpectedHTTPStatusError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", httpErr.StatusCode)
	}
	if n := atomic.LoadInt32(&attempts); n != 1 {
		t.Fatalf("expected 1 request (no retry for 500), got %d", n)
	}
}

func TestFlush_NonRetryable400(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	addTestMetric(t, c)

	err := c.Flush()
	if err == nil {
		t.Fatal("expected error for 400, got nil")
	}
	httpErr, ok := err.(*UnexpectedHTTPStatusError)
	if !ok {
		t.Fatalf("expected *UnexpectedHTTPStatusError, got %T: %v", err, err)
	}
	if httpErr.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", httpErr.StatusCode)
	}
	if n := atomic.LoadInt32(&attempts); n != 1 {
		t.Fatalf("expected 1 request (no retry for 400), got %d", n)
	}
}

func TestFlush_LastFlushAttemptResetOnRetry(t *testing.T) {
	var attempts int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&attempts, 1)
		if n <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	c := newTestClient(t, srv.URL)
	addTestMetric(t, c)

	before := time.Now()
	if err := c.Flush(); err != nil {
		t.Fatalf("expected success, got %v", err)
	}

	// lastFlushAttempt should have been reset during retries,
	// so it should be after the time we captured before Flush.
	if c.lastFlushAttempt.Before(before) {
		t.Fatal("expected lastFlushAttempt to be updated during retry backoff")
	}
	if n := atomic.LoadInt32(&attempts); n != 3 {
		t.Fatalf("expected 3 requests (1 initial + 2 retries), got %d", n)
	}
}
