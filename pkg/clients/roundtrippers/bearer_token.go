package roundtrippers

import (
	"fmt"
	"net/http"
)

type bearerTokenRoundTripper struct {
	token string
	rt    http.RoundTripper
}

// RoundTrip implements http.RoundTripper's interface
func (rt *bearerTokenRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if len(req.Header.Get("Authorization")) == 0 {
		req = cloneRequest(req)
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", rt.token))
	}
	return rt.rt.RoundTrip(req)
}

// NewBearerToken returns an http.RoundTripper that adds the bearer token to a request's header
func NewBearerToken(token string, rt http.RoundTripper) http.RoundTripper {
	return &bearerTokenRoundTripper{token, rt}
}
