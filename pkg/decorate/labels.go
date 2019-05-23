package decorate

import dto "github.com/prometheus/client_model/go"

// Labels is a list of label pairs that need to be added/overwritten on all metrics
type Labels []*dto.LabelPair

// Decorate adds/overwritesa metric labels from its list
func (l Labels) Decorate(mfs []*dto.MetricFamily) {
	for _, fam := range mfs {
		metrics := fam.GetMetric()
		for _, metric := range metrics {
			metric.Label = append(metric.Label, l...)
		}
	}
}

// Name is the name of this decorator
func (Labels) Name() string {
	return "Labels"
}
