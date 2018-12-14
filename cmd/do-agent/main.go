package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/digitalocean/do-agent/internal/log"
	"github.com/prometheus/client_golang/prometheus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGHUP, syscall.SIGINT)
	go func() {
		if sig := <-stop; sig != nil {
			log.Info("caught signal, shutting down: %s", sig.String())
		}
		cancel()
	}()

	os.Args = append(os.Args, additionalParams...)

	// parse all command line flags
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()

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

	w, th := initWriter(ctx)
	d := initDecorator()
	run(ctx, w, th, d, reg)
}
