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

package metrics

// MetricRef is a unique identifier for a metric, provided when the metric is
// registered.  It is used by the related collector when reporting metric values.
type MetricRef interface{}

// Collector is a function which reports current metric values.
type Collector func(r Reporter)

// Registry is a tracked set of metrics.
// When a new metric is added, a unique id is provided to the caller.
// This should be used within a registered collection function to report metrics
// at collection time.
type Registry interface {
	// Register defines a metric collector for custom named metrics, returning
	// a unique id that is used to record metric samples.
	Register(name string, opts ...RegOpt) MetricRef

	// AddCollector adds a collection function to be called to collect metrics.
	// The collector reports the current value of metrics via the Reporter.
	AddCollector(f Collector)

	// Report has all registered collectors report to the given reporter.
	Report(r Reporter)
}

// Reporter defines a metric measurement reporting API.
// Collection functions are periodically called to report metric values.
type Reporter interface {
	Update(ref MetricRef, value float64, labelValues ...string)
}

// NewRegistry returns a new registry.
func NewRegistry() Registry {
	return new(registry)
}

// RegOpt is an option initializer for metric registration.
// Use With* calls to add options.
type RegOpt func(*Definition)

// Definition holds the description of a metric.
type Definition struct {
	Name              string            `json:"name"`
	Type              MetricType        `json:"type"`
	CommonLabels      map[string]string `json:"labels,omitempty"`
	MeasuredLabelKeys []string          `json:"label_keys,omitempty"`
}

// AsType sets the metric type, or else the default is Gauge.
func AsType(t MetricType) RegOpt {
	return func(o *Definition) {
		o.Type = t
	}
}

// WithCommonLabels adds common labels that are used for every measurement on
// the associated metric.
func WithCommonLabels(labels map[string]string) RegOpt {
	return func(o *Definition) {
		if o.CommonLabels == nil {
			o.CommonLabels = labels
			return
		}
		for k, v := range labels {
			o.CommonLabels[k] = v
		}
	}
}

// WithMeasuredLabels adds label keys to the associated metric.  Each time
// the metric is measured, the associated label values must be provided (in
// order!).  If label values are constant for a metric (eg hostname), then
// use WithCommonLabels instead.
func WithMeasuredLabels(labelKeys ...string) RegOpt {
	return func(o *Definition) {
		o.MeasuredLabelKeys = append(o.MeasuredLabelKeys, labelKeys...)
	}
}

type registry struct {
	collectors []Collector
}

func (r *registry) Register(name string, opts ...RegOpt) MetricRef {
	d := &Definition{
		Name: name,
		Type: MetricType_GAUGE,
	}
	for _, o := range opts {
		o(d)
	}

	return d
}

func (r *registry) AddCollector(c Collector) {
	r.collectors = append(r.collectors, c)
}

func (r *registry) Report(reporter Reporter) {
	for _, c := range r.collectors {
		c(reporter)
	}
}
