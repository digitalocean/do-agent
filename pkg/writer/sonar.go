package writer

import (
	"fmt"

	"github.com/digitalocean/do-agent/internal/log"
	"github.com/digitalocean/do-agent/pkg/aggregate"
	"github.com/digitalocean/do-agent/pkg/clients/tsclient"

	"github.com/pkg/errors"
)

var (
	// ErrMetricTooLong is returned when trying to write a metric that exceeds the length limit
	// defined by client.MaxMetricLength
	ErrMetricTooLong = fmt.Errorf("metric length is too long to write")
	// ErrTooManyMetrics is returned when calling Write with too many metrics
	// defined by client.MaxBatchSize
	ErrTooManyMetrics = fmt.Errorf("too many metrics to send")

	// ErrFlushFailure is returned when Flush fails for any reason
	ErrFlushFailure = fmt.Errorf("flush failure")
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
func (s *Sonar) Write(mets []aggregate.MetricWithValue) error {
	if len(mets) > s.client.MaxBatchSize() {
		return errors.Wrap(ErrTooManyMetrics, "cannot write metrics")
	}

	for _, m := range mets {
		lfmEncoded := tsclient.ConvertLFMMapToPrometheusEncodedName(m.LFM)
		if len(lfmEncoded) > s.client.MaxMetricLength() {
			return errors.Wrapf(ErrMetricTooLong, "cannot send metric: %q", lfmEncoded)
		}
		err := s.client.AddMetric(tsclient.NewDefinitionFromMap(m.LFM), m.Value)
		if err != nil {
			return err
		}
	}

	err := s.client.Flush()
	httpError, ok := err.(*tsclient.UnexpectedHTTPStatusError)
	if !s.firstWriteSent && ok && httpError.StatusCode == 429 {
		err = nil
	}
	s.firstWriteSent = true

	if err == nil {
		return nil
	}

	log.Error("failed to flush: %+v", err)
	return ErrFlushFailure
}

// Name is the name of this writer
func (s *Sonar) Name() string {
	return "sonar"
}
