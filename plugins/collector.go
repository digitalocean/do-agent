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

package plugins

import (
	"encoding/json"

	"github.com/digitalocean/do-agent/log"
	"github.com/digitalocean/do-agent/metrics"
)

type initResult struct {
	// Definitions indexed by name.
	Definitions map[string]*metrics.Definition `json:"definitions"`
}

type metricsResult struct {
	// Metric values indexed by name.
	Metrics map[string]*metricValue `json:"metrics"`
}

type metricValue struct {
	Value       float64  `json:"value"`
	LabelValues []string `json:"label_values,omitempty"`
}

// RegisterPluginDir adds a collector to retrieve metrics from external
// plugins.
func RegisterPluginDir(r metrics.Registry, dirPath string) {
	h := NewExternalPluginHandler(dirPath)

	refs := make(map[string]metrics.MetricRef)

	for _, result := range h.ExecuteAll("config") {
		if len(result.Stderr) > 0 {
			log.Errorf("plugin error %q: %s", result.PluginPath, result.Stderr)
		}

		var res initResult
		err := json.Unmarshal(result.Output, &res)
		if err != nil {
			log.Errorf("unable to parse plugin %q: %s", result.PluginPath, err)
			h.RemovePlugin(result.PluginPath)
			continue
		}
		if len(res.Definitions) == 0 {
			log.Debugf("no metric definitions in %q", result.PluginPath)
			h.RemovePlugin(result.PluginPath)
			continue
		}

		for name, d := range res.Definitions {
			ref := r.Register(name, metrics.AsType(d.Type),
				metrics.WithCommonLabels(d.CommonLabels),
				metrics.WithMeasuredLabels(d.MeasuredLabelKeys...))
			refs[name] = ref
		}
	}

	r.AddCollector(func(reporter metrics.Reporter) {
		for _, result := range h.ExecuteAll() {
			if len(result.Stderr) > 0 {
				log.Errorf("plugin error %q: %s", result.PluginPath, result.Stderr)
			}

			var res metricsResult
			err := json.Unmarshal(result.Output, &res)
			if err != nil {
				log.Errorf("unable to parse plugin %q: %s", result.PluginPath, err)
				continue
			}

			for name, m := range res.Metrics {
				ref, ok := refs[name]
				if !ok {
					log.Debugf("undefined metric from plugin %q: %s",
						result.PluginPath, name)
					continue
				}

				reporter.Update(ref, m.Value, m.LabelValues...)
			}
		}
	})
}
