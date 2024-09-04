package main

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/digitalocean/do-agent/internal/log"
)

func main() {
	initConfig()

	if config.debug {
		log.SetLevel(log.LevelDebug)
	}

	if config.syslog {
		if err := log.InitSyslog(); err != nil {
			log.Error("failed to initialize syslog. Using standard logging: %+v", err)
		}
	}

	if err := checkConfig(); err != nil {
		log.Fatal("configuration failure: %+v", err)
	}

	cols := initCollectors()
	reg := prometheus.NewRegistry()
	reg.MustRegister(cols...)

	if config.webListen {
		//Create a secondary registry for local only metrics
		localReg := prometheus.NewRegistry()
		localCols := append(cols, metricWriterDiagnostics)
		localReg.MustRegister(localCols...)
		go func() {
			http.Handle("/", promhttp.HandlerFor(localReg, promhttp.HandlerOpts{}))
			err := http.ListenAndServe(config.webListenAddress, nil)
			if err != nil {
				log.Error("failed to init HTTP listener: %+v", err.Error())
			}
		}()
	}

	w, th := initWriter(metricWriterDiagnostics)
	d := initDecorator()
	aggregateSpecs := initAggregatorSpecs()

	run(w, th, d, reg, aggregateSpecs)
}
