package main

import (
	"time"

	"github.com/digitalocean/do-agent/internal/log"
	"github.com/digitalocean/do-agent/pkg/decorate"
	dto "github.com/prometheus/client_model/go"
)

type metricWriter interface {
	Write(mets []*dto.MetricFamily) error
	Name() string
}

type throttler interface {
	WaitDuration() time.Duration
	Name() string
}

type gatherer interface {
	Gather() ([]*dto.MetricFamily, error)
}

func run(w metricWriter, th throttler, dec decorate.Decorator, g gatherer) {
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
		if err != nil {
			log.Error("failed to send metrics: %v", err)
			return
		}
		log.Debug("stats written in %s", time.Since(start))
	}

	exec()
	for range time.After(th.WaitDuration()) {
		exec()
	}
}
