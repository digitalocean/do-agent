package main

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/digitalocean/do-agent/internal/log"
	"github.com/digitalocean/do-agent/pkg/aggregate"
	"github.com/digitalocean/do-agent/pkg/decorate"
)

const (
	diagnosticMetricName        = "sonar_diagnostic"
	metricWriterDiagnosticsName = "metric_writes"
)

var (
	//ErrAggregationFailed is the error msg for failed aggregation
	ErrAggregationFailed = fmt.Errorf("metric aggregation failed")

	diagnosticMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "",
		Name:      diagnosticMetricName,
		Help:      "do-agent diagnostic information",
	}, []string{"error"})

	metricWriterDiagnostics = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: "",
			Name:      metricWriterDiagnosticsName,
			Help:      "Total successes and failures of metric writers",
		},
		[]string{"writer", "result", "reason"},
	)
)

type metricWriter interface {
	Write(mets []aggregate.MetricWithValue) error
	Name() string
}

type limiter interface {
	WaitDuration() time.Duration
	Name() string
}

type gatherer interface {
	Gather() ([]*dto.MetricFamily, error)
}

func run(w metricWriter, l limiter, dec decorate.Decorator, g gatherer, aggregateSpec map[string][]string) {
	exec := func() {
		start := time.Now()
		mfs, err := g.Gather()
		if err != nil {
			log.Error("failed to gather metrics: %v", err)
			return
		}
		log.Debug("stats collected in %s", time.Since(start))

		start = time.Now()
		dec.Decorate(mfs)
		log.Debug("stats decorated in %s", time.Since(start))

		start = time.Now()
		aggregated, err := aggregate.Aggregate(mfs, aggregateSpec)
		if err != nil {
			log.Error("failed to aggregate metrics: %v", err)
			writeDiagnostics(w, mfs, ErrAggregationFailed)
			return
		}
		log.Debug("stats aggregated in %s", time.Since(start))

		start = time.Now()
		err = w.Write(aggregated)
		if err == nil {
			log.Debug("stats written in %s", time.Since(start))
			return
		}

		log.Error("failed to send metrics: %v", err)
		// After 429 retries the agent has already spent ~45s backing off, so
		// an additional full-cycle sleep would create an unnecessarily long
		// gap (~5 min). Skip it and let the main-loop sleep provide pacing.
		// The diagnostic write may be blocked by the client-side rate limiter;
		// that's acceptable — diagnostics are best-effort.
		writeDiagnostics(w, mfs, err)
	}

	exec()
	for {
		time.Sleep(l.WaitDuration())
		exec()
	}
}

// writeDiagnostics filters all metrics and gathers only the diagnostic information and sends the metrics
// in the event of a write failure
func writeDiagnostics(w metricWriter, mfs []*dto.MetricFamily, err error) {
	diagnosticMetric.WithLabelValues(err.Error()).Inc()
	var diags []*dto.MetricFamily

	for _, mf := range mfs {
		switch mf.GetName() {
		case buildInfoMetricName, diagnosticMetricName:
			diags = append(diags, mf)
		}
	}
	if len(diags) == 0 {
		log.Error("couldn't find any diagnostic information to send, skipping")
		return
	}
	diagnostics, err := aggregate.Aggregate(diags, nil)
	if err != nil {
		log.Error("failed to aggregate diagnostic information: %v", err)
		return
	}

	if err := w.Write(diagnostics); err != nil {
		log.Error("failed to write diagnostic information: %v", err)
	}
}
