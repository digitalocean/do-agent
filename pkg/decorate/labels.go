package decorate

import (
	dto "github.com/prometheus/client_model/go"
)

// LabelAppender is a list of label pairs that need to be added on all metrics
type LabelAppender []*dto.LabelPair

// Decorate adds metric labels from its list
func (l LabelAppender) Decorate(mfs []*dto.MetricFamily) {
	for _, fam := range mfs {
		metrics := fam.GetMetric()
		for _, metric := range metrics {
			metric.Label = append(metric.Label, l...)
		}
	}
}

// Name is the name of this decorator
func (LabelAppender) Name() string {
	return "LabelsAppender"
}
