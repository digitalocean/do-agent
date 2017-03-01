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

// traditionally ticks are messured per second. With modern
// architectures this value may vary.
const ticksPerSecond = 100

type statFunc func() (procfs.Stat, error)

// RegisterCPUMetrics registers CPU related metrics.
func RegisterCPUMetrics(r metrics.Registry, fn statFunc, f Filters) {
	cpu := r.Register("cpu", metrics.WithMeasuredLabels("cpu", "mode"),
		metrics.AsType(metrics.MetricType_COUNTER))
	interrupt := r.Register("intr",
		metrics.AsType(metrics.MetricType_COUNTER))
	contextSwitch := r.Register("context_switches")
	procsBlocked := r.Register("procs_blocked")
	procsRunning := r.Register("procs_running")

	r.AddCollector(func(r metrics.Reporter) {
		stat, err := fn()
		if err != nil {
			log.Debugf("Could not gather cpu metrics: %s", err)
			return
		}

		for _, value := range stat.CPUS {
			if value.CPU == "cpu" {
				continue
			}
			f.UpdateIfIncluded(r, cpu, float64(value.Guest)/ticksPerSecond, value.CPU, "guest")
			f.UpdateIfIncluded(r, cpu, float64(value.GuestNice)/ticksPerSecond, value.CPU, "guestnice")
			f.UpdateIfIncluded(r, cpu, float64(value.Idle)/ticksPerSecond, value.CPU, "idle")
			f.UpdateIfIncluded(r, cpu, float64(value.Iowait)/ticksPerSecond, value.CPU, "iowait")
			f.UpdateIfIncluded(r, cpu, float64(value.Irq)/ticksPerSecond, value.CPU, "irq")
			f.UpdateIfIncluded(r, cpu, float64(value.Nice)/ticksPerSecond, value.CPU, "nice")
			f.UpdateIfIncluded(r, cpu, float64(value.Softirq)/ticksPerSecond, value.CPU, "softirq")
			f.UpdateIfIncluded(r, cpu, float64(value.Steal)/ticksPerSecond, value.CPU, "steal")
			f.UpdateIfIncluded(r, cpu, float64(value.System)/ticksPerSecond, value.CPU, "system")
			f.UpdateIfIncluded(r, cpu, float64(value.User)/ticksPerSecond, value.CPU, "user")
		}

		f.UpdateIfIncluded(r, interrupt, float64(stat.Interrupt))
		f.UpdateIfIncluded(r, contextSwitch, float64(stat.ContextSwitch))
		f.UpdateIfIncluded(r, procsBlocked, float64(stat.ProcessesBlocked))
		f.UpdateIfIncluded(r, procsRunning, float64(stat.ProcessesRunning))
	})
}
