package roundtrippers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	tokenPath        = "testdata/token"
	invalidTokenPath = "testdata/missingToken"
)

func Test_bearerTokenFileRoundTripper_RoundTrip_Happy_Path(t *testing.T) {
	rt := NewBearerTokenFile(tokenPath, http.DefaultTransport)

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

func Test_bearerTokenFileRoundTripper_RoundTrip_Missing_File(t *testing.T) {
	rt := NewBearerTokenFile(invalidTokenPath, http.DefaultTransport)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		return
	}))
	defer ts.Close()

	_, err := rt.RoundTrip(httptest.NewRequest(http.MethodGet, ts.URL, nil))
	if err == nil {
		t.Errorf("Expected error, got none")
	}
}
