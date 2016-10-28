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

package bootstrap

import (
	"fmt"

	"github.com/digitalocean/do-agent/monitoringclient"
)

//Credentials contains do-agent credentials and config required to talk to monitoring
type Credentials struct {
	AppKey    string   `json:"appkey,omitempty"`
	HostUUID  string   `json:"host_uuid,omitempty"`
	Region    string   `json:"region,omitempty"`
	DropletID int64    `json:"droplet_id,omitempty"`
	LocalMACs []string `json:"local_macs,omitempty"`
}

// MetadataReader is the interface a DigitalOcean metadata client should implement
type MetadataReader interface {
	DropletID() (int, error)
	Region() (string, error)
	AuthToken() (string, error)
}

func loadCredentialFromMetadata(md MetadataReader, monitor monitoringclient.MonitoringClient) (*Credentials, error) {
	authToken, err := md.AuthToken()
	if err != nil {
		return nil, err
	}

	appKey, err := monitor.GetAppKey(authToken)
	if err != nil {
		return nil, err
	}

	dropletID, err := md.DropletID()
	if err != nil {
		return nil, err
	}

	region, err := md.Region()
	if err != nil {
		return nil, err
	}

	return &Credentials{
		AppKey:    appKey,
		DropletID: int64(dropletID),
		Region:    region,
	}, nil
}

func loadCredentialWithOverrides(monitor monitoringclient.MonitoringClient, configAppKey string, configDropletID int64, configAuthToken string) (*Credentials, error) {
	appkey := configAppKey
	if configAuthToken != "" {
		sAppKey, err := monitor.GetAppKey(configAuthToken)
		if err != nil {
			return nil, fmt.Errorf("unable to retrieve appkey from monitoring: %s", err.Error())
		}
		appkey = sAppKey
	}

	return &Credentials{
		AppKey:    appkey,
		DropletID: configDropletID,
		Region:    "master",
	}, nil
}

// InitCredentials will read (or create) the credentials file (if running on a droplet)
func InitCredentials(md MetadataReader, monitor monitoringclient.MonitoringClient, configAppKey string, configDropletID int64, configAuthToken string) (*Credentials, error) {
	if configAppKey != "" || configAuthToken != "" {
		return loadCredentialWithOverrides(monitor, configAppKey, configDropletID, configAuthToken)
	}
	return loadCredentialFromMetadata(md, monitor)
}
