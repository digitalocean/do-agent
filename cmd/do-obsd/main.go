package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/digitalocean/do-agent/internal/log"
)

var (
	version    = "dev"
	revision   = "none"
	syslogFlag = flag.Bool("syslog", false, "enable logging to syslog")
	versionFlag = flag.Bool("version", false, "display the version and exit")
)

func main() {
	flag.Parse()

	if *versionFlag {
		fmt.Printf("do-obsd %s (rev: %s)\n", version, revision)
		os.Exit(0)
	}

	if *syslogFlag {
		if err := log.InitSyslog(); err != nil {
			log.Error("failed to initialize syslog: %+v", err)
		}
	}

	log.SetLevel(log.LevelDebug)
	log.Debug("do-obsd: starting version=%s revision=%s pid=%d", version, revision, os.Getpid())

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT)
	sig := <-sigCh

	log.Debug("do-obsd: received %s, shutting down", sig)
}
