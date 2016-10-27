// Copyright 2016 DigitalOcean
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package monitoringclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/digitalocean/do-agent/config"
)

func TestGetAppKeyRejectsNonHTTPS(t *testing.T) {
	requireHTTPS = true
	monitoringClient := &monitoringClient{
		url: "http://insecurelink",
	}
	_, err := monitoringClient.GetAppKey("abc")
	if err != ErrAuthURLNotHTTPS {
		t.Error("unsecure URL accepted")
	}
}

func TestGetAppKey(t *testing.T) {
	requireHTTPS = false
	defer func() { requireHTTPS = true }()

	expectedAppkey := "appkey"
	s := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				fmt.Fprintf(w, "\"%s\"", expectedAppkey)
			}))
	defer s.Close()

	monitoringClient := &monitoringClient{
		url: s.URL,
	}
	appkey, err := monitoringClient.GetAppKey("abc")
	if err != nil {
		t.Fatal(err)
	}
	if appkey != expectedAppkey {
		t.Errorf("got: %s want: %s", appkey, expectedAppkey)
	}
}

func TestAddUserAgentToHTTPRequest(t *testing.T) {
	expectedAgent := "do-agent-" + config.Version()

	r, _ := http.NewRequest("POST", "http://www.digitalocean.com", nil)
	addUserAgentToHTTPRequest(r)

	if r.UserAgent() != expectedAgent {
		t.Errorf("got: %s expected: %s", r.UserAgent(), expectedAgent)
	}
}
