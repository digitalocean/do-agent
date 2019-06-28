package aggregate

import (
	"github.com/pkg/errors"
	dto "github.com/prometheus/client_model/go"

	"github.com/digitalocean/do-agent/pkg/clients/tsclient"
)

// MetricWithValue is a representation of a label formatted metric with a value
type MetricWithValue struct {
	LFM   map[string]string
	Value float64
}

// Aggregate aggregates metric families according to the given aggregate spec.
// A spec with key: {"metricName": "aggregateLabel"} will remove the "aggregateLabel" from all
// "metricName" metric families
func Aggregate(metrics []*dto.MetricFamily, aggregateSpec map[string][]string) ([]MetricWithValue, error) {
	agg := map[string]MetricWithValue{}
	for _, mf := range metrics {
		for _, metric := range mf.Metric {
			var value float64
			switch *mf.Type {
			case dto.MetricType_GAUGE:
				value = *metric.Gauge.Value
			case dto.MetricType_COUNTER:
				value = *metric.Counter.Value
			case dto.MetricType_UNTYPED:
				value = *metric.Untyped.Value
			default:
				// we currently don't support other types of metrics
				continue
			}

			labels := map[string]string{}
			tslbls := make([]string, len(metric.Label)*2)
			for i, label := range metric.Label {
				tslbls[i] = *label.Name
				tslbls[i*2] = *label.Value
				labels[*label.Name] = *label.Value
			}

			def := tsclient.NewDefinition(*mf.Name, tsclient.WithCommonLabels(labels))
			lfm, err := tsclient.GetLFM(def, tslbls)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			lfmDelim, err := tsclient.ParseMetricDelimited(lfm)
			if err != nil {
				return nil, errors.WithStack(err)
			}
			labelsToRemove, ok := aggregateSpec[mf.GetName()]
			if ok {
				// if the metric family is to be aggregated, aggregate away the specified labels
				for _, lbl := range labelsToRemove {
					delete(lfmDelim, lbl)
				}
			}
			key := tsclient.ConvertLFMMapToPrometheusEncodedName(lfmDelim)
			aggregated, ok := agg[key]
			if !ok {
				aggregated.LFM = lfmDelim
			}
			aggregated.Value += value
			agg[key] = aggregated
		}
	}
	squashed := make([]MetricWithValue, 0)
	for _, m := range agg {
		squashed = append(squashed, m)
	}
	return squashed, nil
}
