package writer

import (
	"fmt"

	"github.com/digitalocean/do-agent/pkg/clients/tsclient"

	"github.com/pkg/errors"
	dto "github.com/prometheus/client_model/go"
)

var (
	// ErrMetricTooLong is returned when trying to write a metric that exceeds the length limit
	// defined by client.MaxMetricLength
	ErrMetricTooLong = fmt.Errorf("metric length is too long to write")
	// ErrTooManyMetrics is returned when calling Write with too many metrics
	// defined by client.MaxBatchSize
	ErrTooManyMetrics = fmt.Errorf("too many metrics to send")
)

// Sonar writes metrics to DigitalOcean sonar
type Sonar struct {
	client         tsclient.Client
	firstWriteSent bool
}

// NewSonar creates a new Sonar writer
func NewSonar(client tsclient.Client) *Sonar {
	return &Sonar{
		client:         client,
		firstWriteSent: false,
	}
}

// Write writes the metrics to Sonar and returns the amount of time to wait
// before the next write
func (s *Sonar) Write(mets []*dto.MetricFamily) error {
	if len(mets) > s.client.MaxBatchSize() {
		return errors.Wrap(ErrTooManyMetrics, "cannot write metrics")
	}

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
			tslbls := make([]string, len(metric.Label)*2)
			for i, label := range metric.Label {
				tslbls[i] = *label.Name
				tslbls[i*2] = *label.Value
				labels[*label.Name] = *label.Value
			}

			def := tsclient.NewDefinition(*mf.Name, tsclient.WithCommonLabels(labels))
			lfm, err := tsclient.GetLFM(def, tslbls)
			if err != nil {
				return errors.WithStack(err)
			}
			if len(lfm) > s.client.MaxMetricLength() {
				return errors.Wrapf(ErrMetricTooLong, "cannot send metric: %q", metric.String())
			}

			err = s.client.AddMetric(def, value)
			if err != nil {
				return err
			}

		}

	}
	err := s.client.Flush()
	httpError, ok := err.(*tsclient.UnexpectedHTTPStatusError)
	if !s.firstWriteSent && ok && httpError.StatusCode == 429 {
		err = nil
	}
	s.firstWriteSent = true
	return err
}

// Name is the name of this writer
func (s *Sonar) Name() string {
	return "sonar"
}
