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

type stubProcer struct {
	NewProcProcResultProcProcs []procfs.ProcProc
	NewProcProcResultErr       error
}

func (s *stubProcer) NewProcProc() ([]procfs.ProcProc, error) {
	return s.NewProcProcResultProcProcs, s.NewProcProcResultErr
}

// Verify that the stubProcProc implements the procfs.Procer interface.
var _ procfs.Procer = (*stubProcer)(nil)

func TestRegisterProcessMetrics(t *testing.T) {
	p := &stubProcer{}
	p.NewProcProcResultErr = nil
	p.NewProcProcResultProcProcs = []procfs.ProcProc{
		procfs.ProcProc{
			PID:            int(1),
			CPUUsage:       float64(2),
			ResidentMemory: int(3),
			VirtualMemory:  int(4),
			Comm:           "foo",
			CmdLine:        []string{"a", "b"},
		},
		procfs.ProcProc{
			PID:            int(2),
			CPUUsage:       float64(3),
			ResidentMemory: int(2),
			VirtualMemory:  int(1),
			Comm:           "foo",
			CmdLine:        []string{"a", "b"},
		},
	}

	expectedNames := []string{
		"process_memory",
		"process_cpu",
	}

	var actualNames []string

	r := &stubRegistry{}

	RegisterProcessMetrics(r, p.NewProcProc)

	for i := range r.RegisterNameOpts {
		actualNames = append(actualNames, r.RegisterNameOpts[i].Name)
	}

	testForMetricNames(t, expectedNames, actualNames)

	if r.AddCollectorFunc == nil {
		t.Error("expected collector function, found none")
	}
}
