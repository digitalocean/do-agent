package main

import (
	"github.com/digitalocean/do-agent/internal/log"

	"github.com/prometheus/client_golang/prometheus"
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

	w, th := initWriter(reg)
	d := initDecorator()
	run(w, th, d, reg)
}
