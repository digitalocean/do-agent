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
	totalCPUTime       float64
	userCPUTime        float64
	kernelCPUTime      float64
	childUserCPUTime   float64
	childKernelCPUTime float64
	startTimeCPUTime   float64
	totalMemory        float64
}

type procprocFunc func() ([]procfs.ProcProc, error)

// RegisterProcessMetrics registers process metrics.
func RegisterProcessMetrics(r metrics.Registry, fn procprocFunc) {
	memory := r.Register(processSystem+"_memory",
		metrics.WithMeasuredLabels("process"))
	cpu := r.Register(processSystem+"_cpu",
		metrics.WithMeasuredLabels("process"))

	ucpu := r.Register(processSystem+"_cpu_user",
		metrics.WithMeasuredLabels("process"))
	scpu := r.Register(processSystem+"_cpu_kernel",
		metrics.WithMeasuredLabels("process"))
	cucpu := r.Register(processSystem+"_cpu_child_user",
		metrics.WithMeasuredLabels("process"))
	cscpu := r.Register(processSystem+"_cpu_child_kernel",
		metrics.WithMeasuredLabels("process"))
	startcpu := r.Register(processSystem+"_cpu_start",
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
				value.totalCPUTime += proc.CPUTime
				value.userCPUTime += proc.UserCPUTime
				value.kernelCPUTime += proc.KernelCPUTime
				value.childUserCPUTime += proc.ChildUserCPUTime
				value.childKernelCPUTime += proc.ChildKernelCPUTime
				value.totalMemory += float64(proc.ResidentMemory)
			} else {
				m[proc.Comm] = &process{
					totalCPUTime: proc.CPUTime,
					totalMemory:  float64(proc.ResidentMemory),

					userCPUTime:        proc.UserCPUTime,
					kernelCPUTime:      proc.KernelCPUTime,
					childUserCPUTime:   proc.ChildUserCPUTime,
					childKernelCPUTime: proc.ChildKernelCPUTime,
					startTimeCPUTime:   proc.StartTimeCPUTime,
				}
			}
		}

		for key, value := range m {
			r.Update(memory, value.totalMemory, key)
			r.Update(cpu, value.totalCPUTime, key)

			r.Update(ucpu, value.userCPUTime, key)
			r.Update(scpu, value.kernelCPUTime, key)
			r.Update(cucpu, value.childUserCPUTime, key)
			r.Update(cscpu, value.childKernelCPUTime, key)
			r.Update(startcpu, value.startTimeCPUTime, key)
		}
	})
}
