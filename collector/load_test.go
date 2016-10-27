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

type stubLoader struct {
	NewLoadResultLoad procfs.Load
	NewLoadResultErr  error
}

func (s *stubLoader) NewLoad() (procfs.Load, error) {
	return s.NewLoadResultLoad, s.NewLoadResultErr
}

// Verify that the stubMounter implements the procfs.Mounter interface.
var _ procfs.Loader = (*stubLoader)(nil)

func TestRegisterLoadMetrics(t *testing.T) {
	l := &stubLoader{}
	l.NewLoadResultErr = nil
	l.NewLoadResultLoad = procfs.Load{
		Load1:        float64(1),
		Load5:        float64(2),
		Load15:       float64(3),
		RunningProcs: uint64(4),
		TotalProcs:   uint64(5),
		LastPIDUsed:  uint64(6),
	}

	expectedNames := []string{
		"load1",
		"load5",
		"load15",
	}

	var actualNames []string

	r := &stubRegistry{}

	RegisterLoadMetrics(r, l.NewLoad)

	for i := range r.RegisterNameOpts {
		actualNames = append(actualNames, r.RegisterNameOpts[i].Name)
	}

	testForMetricNames(t, expectedNames, actualNames)

	if r.AddCollectorFunc == nil {
		t.Error("expected collector function, found none")
	}
}
