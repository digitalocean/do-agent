package tsclient

import "fmt"

// UnexpectedHTTPStatusError is returned when an unexpected HTTP status is
// returned when making a registry api call.
type UnexpectedHTTPStatusError struct {
	StatusCode int
}

// Error returns the error string
func (e *UnexpectedHTTPStatusError) Error() string {
	return fmt.Sprintf("received unexpected HTTP status: %d", e.StatusCode)
}

// ErrSendTooFrequent happens if the client attempts to send metrics faster
// than what the server requested
var ErrSendTooFrequent = fmt.Errorf("metrics sent faster than server requested")

// ErrFlushTooFrequent happens if the client attempts to flush metrics faster
// than what the server requested
var ErrFlushTooFrequent = fmt.Errorf("metrics sent faster than server requested")

// ErrCircuitBreaker happens when there are many back to back failures sending metrics, the client will
// deliberately fail in order to reduce network load on the server
var ErrCircuitBreaker = fmt.Errorf("circuit breaker is open; deliberately failing due to exponential backoff")

// ErrLabelMissmatch happens if the number of supplied labels is incorrect
var ErrLabelMissmatch = fmt.Errorf("unexpected number of labels")
