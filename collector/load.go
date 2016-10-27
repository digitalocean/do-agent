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

type loadFunc func() (procfs.Load, error)

// RegisterLoadMetrics registers system load related metrics.
func RegisterLoadMetrics(r metrics.Registry, fn loadFunc) {
	load1 := r.Register("load1")
	load5 := r.Register("load5")
	load15 := r.Register("load15")

	r.AddCollector(func(r metrics.Reporter) {
		loads, err := fn()
		if err != nil {
			log.Debugf("couldn't get load: %s", err)
			return
		}
		r.Update(load1, loads.Load1)
		r.Update(load5, loads.Load5)
		r.Update(load15, loads.Load15)
	})
}
