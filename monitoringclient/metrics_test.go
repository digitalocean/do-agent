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
	"strconv"
	"testing"

	"github.com/digitalocean/do-agent/metrics"
)

func TestNewMetricsClientDroplet(t *testing.T) {
	expAppKey := "xxx"
	expDroplet := int64(123)
	expRegion := "miami"
	expURL := "https://miami.sonar.digitalocean.com"

	mcd := newMetricsClientDroplet(expAppKey, expDroplet, expRegion, "")
	if mcd.url != expURL {
		t.Errorf("got: %s expected: %s", mcd.url, expURL)
	}
	if mcd.appKey != expAppKey {
		t.Errorf("got: %s expected: %s", mcd.appKey, expAppKey)
	}
	if mcd.dropletID != expDroplet {
		t.Errorf("got: %d expected: %d", mcd.dropletID, expDroplet)
	}
}

func TestNewMetricsClientDropletNonHTTPS(t *testing.T) {
	requireHTTPS = true
	monitoringClient := &monitoringMetricsClientDroplet{
		url: "http://insecurelink",
		r:   metrics.NewRegistry(),
	}
	_, err := monitoringClient.SendMetrics()
	if err != ErrAuthURLNotHTTPS {
		t.Error("unsecure URL accepted")
	}
}

func TestDropletSendMetrics(t *testing.T) {
	requireHTTPS = false
	defer func() { requireHTTPS = true }()

	expectedAppkey := "appkey"
	expectedDropletID := int64(123456789)
	expectedURLPath := fmt.Sprintf("/v1/metrics/droplet_id/%d", expectedDropletID)
	expectedPushInterval := 52

	s := httptest.NewServer(
		http.HandlerFunc(
			func(w http.ResponseWriter, r *http.Request) {
				if r.URL.String() != expectedURLPath {
					w.WriteHeader(http.StatusNotFound)
				} else {
					w.Header().Set(pushIntervalHeaderKey,
						strconv.FormatInt(int64(expectedPushInterval), 10))
					w.WriteHeader(http.StatusAccepted)
				}
			}))
	defer s.Close()

	monitoringClient := &monitoringMetricsClientDroplet{
		url:       s.URL,
		appKey:    expectedAppkey,
		dropletID: expectedDropletID,
		r:         metrics.NewRegistry(),
	}

	actualPushInterval, err := monitoringClient.SendMetrics()
	if err != nil {
		t.Fatal(err)
	}
	if actualPushInterval != expectedPushInterval {
		t.Errorf("want %v got %v", expectedPushInterval, actualPushInterval)
	}
}

func TestRandomizedPutInterval(t *testing.T) {
	for i := 0; i < 100; i++ {
		j := randomizedPushInterval()
		if j < defaultPushInterval+jitterMin {
			t.Fatalf("interval too short: %d", j)
		}
		if j > defaultPushInterval+jitterMax {
			t.Fatalf("interval too long: %d", j)
		}
	}
}
