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

type stubMounter struct {
	NewMountResultMounts []procfs.Mount
	NewMountResultErr    error
}

func (s *stubMounter) NewMount() ([]procfs.Mount, error) {
	return s.NewMountResultMounts, s.NewMountResultErr
}

// Verify that the stubMounter implements the procfs.Mounter interface.
var _ procfs.Mounter = (*stubMounter)(nil)

func TestRegisterFSMetrics(t *testing.T) {
	m := &stubMounter{}
	m.NewMountResultErr = nil
	m.NewMountResultMounts = []procfs.Mount{
		procfs.Mount{
			Device:     "rootfs",
			MountPoint: "/",
			FSType:     "shoes",
		},
	}

	expectedNames := []string{
		"filesystem_avail",
		"filesystem_files",
		"filesystem_files_free",
		"filesystem_free",
		"filesystem_size",
	}

	var actualNames []string

	r := &stubRegistry{}
	f := Filters{IncludeAll: true}
	RegisterFSMetrics(r, m.NewMount, f)

	for i := range r.RegisterNameOpts {
		actualNames = append(actualNames, r.RegisterNameOpts[i].Name)
	}

	testForMetricNames(t, expectedNames, actualNames)

	if r.AddCollectorFunc == nil {
		t.Error("expected collector function, found none")
	}
}
