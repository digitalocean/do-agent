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
	"bytes"
	"encoding/json"
	"io"

	"github.com/digitalocean/do-agent/log"
	"github.com/digitalocean/do-agent/metrics"
)

// NewExportReporter returns a new reporter for use in plugins when exporting
// metrics.
func NewExportReporter() *ExportReporter {
	return &ExportReporter{
		defs: &initResult{
			Definitions: make(map[string]*metrics.Definition),
		},
		metrics: &metricsResult{
			Metrics: make(map[string]*metricValue),
		},
	}
}

// ExportReporter is a metrics.Reporter implementation which can serialize
// results for a collection plugin.
type ExportReporter struct {
	defs    *initResult
	metrics *metricsResult
}

var _ metrics.Reporter = &ExportReporter{}

// Update handles a metric update.
func (r *ExportReporter) Update(id metrics.MetricRef, value float64,
	labelValues ...string) {

	def, ok := id.(*metrics.Definition)
	if !ok {
		log.Debugf("unknown metric: %d", id)
		return
	}
	if _, ok := r.defs.Definitions[def.Name]; !ok {
		r.defs.Definitions[def.Name] = def
	}

	r.metrics.Metrics[def.Name] = &metricValue{
		Value:       value,
		LabelValues: labelValues,
	}
}

func (r *ExportReporter) Write(w io.Writer, writeConfig bool) error {
	var data interface{}
	if writeConfig {
		data = r.defs
	} else {
		data = r.metrics
	}

	buf, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	_, err = io.Copy(w, bytes.NewBuffer(buf))
	return err
}
