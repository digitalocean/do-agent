package writer

import (
	"github.com/digitalocean/do-agent/pkg/clients/tsclient"
	dto "github.com/prometheus/client_model/go"
)

// Sonar writes metrics to DigitalOcean sonar
type Sonar struct {
	client tsclient.Client
}

// NewSonar creates a new Sonar writer
func NewSonar(client tsclient.Client) *Sonar {
	return &Sonar{
		client: client,
	}
}

// Write writes the metrics to Sonar and returns the amount of time to wait
// before the next write
func (s *Sonar) Write(mets []*dto.MetricFamily) error {
	for _, mf := range mets {
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
				// FIXME -- expand this to support other types
				continue
			}

			labels := map[string]string{}
			for _, label := range metric.Label {
				labels[*label.Name] = *label.Value
			}

			s.client.AddMetric(
				tsclient.NewDefinition(*mf.Name, tsclient.WithCommonLabels(labels)),
				value)

		}

	}

	return s.client.Flush()
}

// Name is the name of this writer
func (s *Sonar) Name() string {
	return "sonar"
}
