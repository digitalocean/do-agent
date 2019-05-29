package main

import (
	"github.com/pkg/errors"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/digitalocean/do-agent/internal/log"
	"github.com/digitalocean/do-agent/pkg/decorate"
)

const (
	diagnosticMetricName = "sonar_diagnostic"
)

var (
	diagnosticMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "",
		Name:      diagnosticMetricName,
		Help:      "do-agent diagnostic information",
	}, []string{"error"})
)

type metricWriter interface {
	Write(mets []*dto.MetricFamily) error
	Name() string
}

type limiter interface {
	WaitDuration() time.Duration
	Name() string
}

type gatherer interface {
	Gather() ([]*dto.MetricFamily, error)
}

func run(w metricWriter, l limiter, dec decorate.Decorator, g gatherer) {
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

		err = w.Write(mfs)
		if err == nil {
			log.Debug("stats written in %s", time.Since(start))
			return
		}

		log.Error("failed to send metrics: %v", err)
		// don't send again immediately or it will fail for sending too frequently
		// first sleep for the wait duration and then send diagnostic information
		time.Sleep(l.WaitDuration())
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
	diagnosticMetric.WithLabelValues(errors.Cause(err).Error()).Inc()
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

	if err := w.Write(diags); err != nil {
		log.Error("failed to write diagnostic information: %v", err)
	}
}
