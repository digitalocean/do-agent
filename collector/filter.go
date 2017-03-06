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
	"regexp"
	"strings"

	"github.com/digitalocean/do-agent/log"
	"github.com/digitalocean/do-agent/metrics"
)

// Filters is used to limit collection of metrics.
type Filters struct {
	IncludeAll bool
	Regexps    []*regexp.Regexp
}

// UpdateIfIncluded call r.Update if the metric should be included.
func (f *Filters) UpdateIfIncluded(r metrics.Reporter, ref metrics.MetricRef, value float64, labelValues ...string) {
	def, ok := ref.(*metrics.Definition)
	if !ok {
		log.Debugf("unknown metric: %d", ref)
		return
	}

	l := def.Name
	if len(labelValues) > 0 {
		l += "_" + strings.Join(labelValues, "_")
	}

	if f.IncludeAll {
		r.Update(ref, value, labelValues...)
		log.Debugf("(+) included via catch all: %v", l)
		return
	}

	for _, e := range f.Regexps {
		if e.MatchString(l) {
			r.Update(ref, value, labelValues...)
			log.Debugf("(+) included via regex: %v", l)
			return
		}
	}

	log.Debugf("(-) excluded: %v", l)
}
