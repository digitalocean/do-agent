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
	"errors"
	"testing"
)

type stubMetadataClient struct {
	DropletIDMethod func() (int, error)
	RegionMethod    func() (string, error)
	AuthTokenMethod func() (string, error)
}

func (s *stubMetadataClient) DropletID() (int, error)    { return s.DropletIDMethod() }
func (s *stubMetadataClient) Region() (string, error)    { return s.RegionMethod() }
func (s *stubMetadataClient) AuthToken() (string, error) { return s.AuthTokenMethod() }

var _ MetadataReader = (*stubMetadataClient)(nil)

type stubMonitoringClient struct {
	getAppKeyMethod        func(string) (string, error)
	registerHostUUIDMethod func(string, string) error
}

func (s *stubMonitoringClient) GetAppKey(authToken string) (string, error) {
	return s.getAppKeyMethod(authToken)
}

func (s *stubMonitoringClient) RegisterHostUUID(appKey, hostUUID string) error {
	return s.registerHostUUIDMethod(appKey, hostUUID)
}

func TestDropletBootstrap(t *testing.T) {

	expectedDropletID := int(999)
	expectedRegion := "test1"
	expectedAuthToken := "authtestabc"

	md := &stubMetadataClient{
		AuthTokenMethod: func() (string, error) { return expectedAuthToken, nil },
		DropletIDMethod: func() (int, error) { return expectedDropletID, nil },
		RegionMethod:    func() (string, error) { return expectedRegion, nil },
	}

	expectedAppKey := "testappkey"

	monitoringClient := struct{ stubMonitoringClient }{}
	monitoringClient.getAppKeyMethod = func(authToken string) (string, error) {
		if authToken == expectedAuthToken {
			return expectedAppKey, nil
		}
		return "", errors.New("Auth token invalid")
	}

	credentials, err := InitCredentials(md, &monitoringClient, "", 0, "")
	if err != nil {
		t.Fatal(err)
	}

	checkCredentials(t, credentials, 0, "", int64(expectedDropletID), expectedRegion, expectedAppKey)

	credentials2, err := InitCredentials(md, &monitoringClient, "", 0, "")
	if err != nil {
		t.Fatal(err)
	}

	checkCredentials(t, credentials2, 0, "", credentials.DropletID, credentials.Region, credentials.AppKey)
}

func TestDropletBootstrapWithOverides(t *testing.T) {

	expectedDropletID := int(999)
	expectedDropletIDOverride := int64(0)
	expectedRegion := "master"
	expectedAuthToken := "authy"

	md := struct{ stubMetadataClient }{}
	md.AuthTokenMethod = func() (string, error) { return expectedAuthToken, nil }
	md.DropletIDMethod = func() (int, error) { return expectedDropletID, nil }
	md.RegionMethod = func() (string, error) { return expectedRegion, nil }

	expectedAppKey := "testappkey"
	expectedAppKeyOverride := "authy"

	monitoringClient := struct{ stubMonitoringClient }{}
	monitoringClient.getAppKeyMethod = func(authToken string) (string, error) {
		if authToken == expectedAuthToken {
			return expectedAppKey, nil
		}
		return "", errors.New("Auth token invalid")
	}

	// Configuring auth token should bypass the metadata service
	credentials, err := InitCredentials(&md, &monitoringClient, "", int64(expectedDropletID), "authy")
	if err != nil {
		t.Fatal(err)
	}

	checkCredentials(t, credentials, 0, "", int64(expectedDropletID), expectedRegion, expectedAppKey)

	// Configure app key should bypass metadata service and sonar service
	credentials2, err := InitCredentials(&md, &monitoringClient, expectedAppKeyOverride, 0, "")
	if err != nil {
		t.Fatal(err)
	}

	checkCredentials(t, credentials2, 0, "", expectedDropletIDOverride, expectedRegion, expectedAppKeyOverride)
}

func checkCredentials(t *testing.T, c *Credentials, eMacsCount int, eHostUUID string, eDropletID int64, eRegion string, eAppKey string) {
	if len(c.LocalMACs) != eMacsCount {
		t.Errorf("LocalMacs: want %d got %d", eMacsCount, len(c.LocalMACs))
	}
	if c.HostUUID != eHostUUID {
		t.Errorf("HostUUID want %s, got %s", eHostUUID, c.HostUUID)
	}
	if c.DropletID != eDropletID {
		t.Errorf("DropletID want %d got %d", eDropletID, c.DropletID)
	}
	if c.Region != eRegion {
		t.Errorf("Region want %s got %s", eRegion, c.Region)
	}
	if c.AppKey != eAppKey {
		t.Errorf("AppKey want %s got %s", eAppKey, c.AppKey)
	}
}
