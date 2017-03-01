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
	"github.com/digitalocean/do-agent/log"
	"github.com/digitalocean/do-agent/metrics"
	"github.com/digitalocean/do-agent/procfs"
)

const processSystem = "process"

type process struct {
	totalCPUUtilization float64
	totalMemory         float64
}

type procprocFunc func() ([]procfs.ProcProc, error)

// RegisterProcessMetrics registers process metrics.
func RegisterProcessMetrics(r metrics.Registry, fn procprocFunc, f Filters) {
	memory := r.Register(processSystem+"_memory",
		metrics.WithMeasuredLabels("process"))
	cpu := r.Register(processSystem+"_cpu",
		metrics.WithMeasuredLabels("process"))

	r.AddCollector(func(r metrics.Reporter) {
		procs, err := fn()
		if err != nil {
			log.Debugf("couldn't get processes: %s", err)
			return
		}

		m := make(map[string]*process)
		for _, proc := range procs {
			if value, ok := m[proc.Comm]; ok {
				value.totalCPUUtilization += proc.CPUUtilization
				value.totalMemory += float64(proc.ResidentMemory)
			} else {
				m[proc.Comm] = &process{
					totalCPUUtilization: proc.CPUUtilization,
					totalMemory:         float64(proc.ResidentMemory),
				}
			}
		}

		for key, value := range m {
			f.UpdateIfIncluded(r, memory, value.totalMemory, key)
			f.UpdateIfIncluded(r, cpu, value.totalCPUUtilization, key)
		}
	})
}
