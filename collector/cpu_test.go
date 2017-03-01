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

type stubStater struct {
	NewStatResultStat procfs.Stat
	NewStatResultErr  error
}

func (s *stubStater) NewStat() (procfs.Stat, error) { return s.NewStatResultStat, s.NewStatResultErr }

// Verify that the stubStater implements the procfs.Stater interface.
var _ procfs.Stater = (*stubStater)(nil)

func TestRegisterCPUMetrics(t *testing.T) {
	stat := &stubStater{}
	stat.NewStatResultErr = nil
	stat.NewStatResultStat = procfs.Stat{
		CPUS: []procfs.CPU{
			procfs.CPU{
				CPU:       "cpu1",
				User:      uint64(1),
				Nice:      uint64(2),
				System:    uint64(3),
				Idle:      uint64(4),
				Iowait:    uint64(5),
				Irq:       uint64(6),
				Softirq:   uint64(7),
				Steal:     uint64(8),
				Guest:     uint64(9),
				GuestNice: uint64(10),
			},
		},
		Interrupt:        uint64(1),
		ContextSwitch:    uint64(2),
		Processes:        uint64(3),
		ProcessesRunning: uint64(4),
		ProcessesBlocked: uint64(5),
	}

	expectedNames := []string{
		"cpu",
		"intr",
		"context_switches",
		"procs_blocked",
		"procs_running",
	}

	var actualNames []string

	r := &stubRegistry{}
	f := Filters{IncludeAll: true}
	RegisterCPUMetrics(r, stat.NewStat, f)

	for i := range r.RegisterNameOpts {
		actualNames = append(actualNames, r.RegisterNameOpts[i].Name)
	}

	testForMetricNames(t, expectedNames, actualNames)

	if r.AddCollectorFunc == nil {
		t.Error("expected collector function, found none")
	}
}
