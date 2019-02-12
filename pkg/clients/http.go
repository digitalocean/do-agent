package clients

import (
	"net"
	"net/http"
	"time"

	"github.com/digitalocean/do-agent/internal/log"
)

// HTTPClient is can make HTTP requests
type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

// NewHTTP creates a new HTTP client with the provided timeout
func NewHTTP(timeout time.Duration) *http.Client {
	return &http.Client{
		Timeout: timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: timeout,
			}).DialContext,
			TLSHandshakeTimeout:   timeout,
			ResponseHeaderTimeout: timeout,
			DisableKeepAlives:     true,
		},
	}
}

// FakeHTTPClient is used for testing
type FakeHTTPClient struct {
	DoFunc func(*http.Request) (*http.Response, error)
}

// Do an HTTP request for testing
func (c *FakeHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if c.DoFunc != nil {
		return c.DoFunc(req)
	}
	return nil, nil
}

// NewDebug creates a new DebugHTTPClient
func NewDebug(timeout time.Duration) *DebugHTTPClient {
	return &DebugHTTPClient{NewHTTP(timeout)}
}

// DebugHTTPClient is an *http.Client that prints Headers and Body to log
type DebugHTTPClient struct {
	*http.Client
}

// Do sends the http request and logs headers and body to DEBUG
func (c *DebugHTTPClient) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.Client.Do(req)
	if err != nil {
		return nil, err
	}

	log.Debug("%T: HTTP %s %s [%d %s]", c, req.Method, req.URL, resp.StatusCode, resp.Status)

	return resp, err
}
