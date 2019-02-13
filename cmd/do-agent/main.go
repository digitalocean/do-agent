package main

import (
	"os"

	"github.com/digitalocean/do-agent/internal/flags"
	"github.com/digitalocean/do-agent/internal/log"

	"github.com/prometheus/client_golang/prometheus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	os.Args = append(os.Args, additionalParams...)

	// read flags from cli directly first so we have access to them
	flags.Init(os.Args[1:])

	// parse all command line flags which are defined across the app
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

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

	w, th := initWriter()
	d := initDecorator()
	run(w, th, d, reg)
}
