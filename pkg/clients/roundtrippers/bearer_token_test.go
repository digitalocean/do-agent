package roundtrippers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	token = "test-token-value"
)

func TestBearerTokenRoundTripper_RoundTrip_Happy_Path(t *testing.T) {
	rt := NewBearerToken(token, http.DefaultTransport)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expectedHeader := fmt.Sprintf("Bearer %s", token)
		if r.Header.Get("Authorization") != expectedHeader {
			t.Errorf("Header.Authorization = %s, want %s", r.Header.Get("Authorization"), expectedHeader)
		}
	}))
	defer ts.Close()

	_, err := rt.RoundTrip(httptest.NewRequest(http.MethodGet, ts.URL, nil))
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
}
