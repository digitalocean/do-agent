package roundtrippers

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type bearerTokenFileRoundTripper struct {
	tokenFile string
	rt        http.RoundTripper
}

// RoundTrip implements http.RoundTripper's interface
func (rt *bearerTokenFileRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	t, err := ioutil.ReadFile(rt.tokenFile)
	if err != nil {
		return nil, fmt.Errorf("unable to read bearer token file %s: %s", rt.tokenFile, err)
	}

	token := strings.TrimSpace(string(t))

	req = cloneRequest(req)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))

	return rt.rt.RoundTrip(req)
}

// NewBearerTokenFile returns an http.RoundTripper that adds the bearer token from a file to a request's header
func NewBearerTokenFile(tokenFile string, rt http.RoundTripper) http.RoundTripper {
	return &bearerTokenFileRoundTripper{tokenFile, rt}
}
