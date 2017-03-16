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
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/digitalocean/do-agent/config"
	"github.com/digitalocean/do-agent/log"
)

const (
	// AuthURL is the address to the Sonar authentication server
	AuthURL = "https://sonar.digitalocean.com"

	// MetadataURL is the address to the metadata service
	MetadataURL = "http://169.254.169.254"

	userAgentHeader = "User-Agent"
)

var (
	// ErrAuthURLNotHTTPS occurs if we connect over http rather than https
	ErrAuthURLNotHTTPS = errors.New("Sonar URL not HTTPS")

	// Allow unittests to disable this
	requireHTTPS = true
)

//MonitoringClient interface describes available sonar API calls
type MonitoringClient interface {
	GetAppKey(string) (string, error)
}

type monitoringClient struct {
	url string
}

// GetAppKey retrieves the appkey from the sonar service.
func (s *monitoringClient) GetAppKey(authToken string) (string, error) {
	err := httpsCheck(s.url)
	if err != nil {
		return "", err
	}

	hc := http.Client{
		Timeout: httpTimeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout: httpTimeout,
			}).Dial,
			TLSHandshakeTimeout:   httpTimeout,
			ResponseHeaderTimeout: httpTimeout,
			DisableKeepAlives:     true,
		},
	}

	req, err := http.NewRequest("GET", s.url+"/v1/appkey/droplet-auth-token", nil)
	if err != nil {
		errMsg := "DigitalOcean sonar service unreachable: %s"
		log.Errorf(errMsg, err)
		return "", err
	}

	addUserAgentToHTTPRequest(req)
	req.Header.Add("Authorization", "DOMETADATA "+authToken)

	resp, err := hc.Do(req)
	if err != nil {
		return "", fmt.Errorf("DigitalOcean sonar service unreachable: %s", err.Error())
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("DigitalOcean sonar service returned unexpected status: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Debugf("Error parsing value from http request: %s", err)
		return "", err
	}

	var appKey string
	err = json.Unmarshal(body, &appKey)
	if err != nil {
		log.Errorf("Failed to unmarshall appKey: %s", err)
		return "", err
	}

	return appKey, nil
}

// NewClient creates a new monitoring client
func NewClient(configURL string) MonitoringClient {
	if configURL != AuthURL {
		requireHTTPS = false
		log.Debugf("HTTPS requirement not enforced for overridden url: %s", configURL)
		return &monitoringClient{
			url: configURL,
		}
	}

	return &monitoringClient{
		url: AuthURL,
	}
}

func httpsCheck(url string) error {
	if !strings.HasPrefix(strings.ToLower(url), "https://") && requireHTTPS {
		return ErrAuthURLNotHTTPS
	}
	return nil
}

// addUserAgentToHTTPRequest adds sonar agent label with agent version
// number to the HTTP user agent header
func addUserAgentToHTTPRequest(req *http.Request) {
	req.Header.Add(userAgentHeader, "do-agent-"+config.Version())
}
