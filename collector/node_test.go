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

package collector

import (
	"testing"

	"github.com/digitalocean/do-agent/procfs"
)

type stubOSReleaser struct {
	NewOSReleaseResultOSRelease procfs.OSRelease
	NewOSReleaseResultErr       error
}

func (s *stubOSReleaser) NewOSRelease() (procfs.OSRelease, error) {
	return s.NewOSReleaseResultOSRelease, s.NewOSReleaseResultErr
}

// Verify that the stubOSReleaser implements the procfs.OSReleaser interface.
var _ procfs.OSReleaser = (*stubOSReleaser)(nil)

func TestRegisterNodeMetrics(t *testing.T) {
	o := &stubOSReleaser{}
	o.NewOSReleaseResultErr = nil
	o.NewOSReleaseResultOSRelease = procfs.OSRelease("lingus")

	expectedNames := []string{"node_info"}

	var actualNames []string

	r := &stubRegistry{}

	RegisterNodeMetrics(r, o.NewOSRelease)

	for i := range r.RegisterNameOpts {
		actualNames = append(actualNames, r.RegisterNameOpts[i].Name)
	}

	testForMetricNames(t, expectedNames, actualNames)

	if r.AddCollectorFunc == nil {
		t.Error("expected collector function, found none")
	}
}
