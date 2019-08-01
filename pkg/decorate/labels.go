package decorate

import (
	"strings"

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

// NamespaceLabelRequired filters out metrics in a given namespace that do not contain a given non-empty label
// It also strips extra labels
type NamespaceLabelRequired struct {
	Namespace string
	Labels    map[string]bool
}

// Decorate by dropping metrics that in a namespace that are either missing a label, or the label value is empty
func (n *NamespaceLabelRequired) Decorate(mfs []*dto.MetricFamily) {
	for _, fam := range mfs {
		if strings.HasPrefix(fam.GetName(), n.Namespace) {
			var list []*dto.Metric
			for _, metric := range fam.GetMetric() {
				if keep, ok := n.hasLabels(metric.Label); ok {
					metric.Label = keep
					list = append(list, metric)
				}
			}

			fam.Metric = list
		}
	}
}

func (n *NamespaceLabelRequired) hasLabels(labels []*dto.LabelPair) ([]*dto.LabelPair, bool) {
	var keep []*dto.LabelPair
	for _, l := range labels {
		if n.Labels[*l.Name] {
			if *l.Value == "" {
				return nil, false
			}

			keep = append(keep, l)
		}
	}

	return keep, true
}

// Name is the name of this decorator
func (*NamespaceLabelRequired) Name() string {
	return "NamespaceLabelRequired"
}
