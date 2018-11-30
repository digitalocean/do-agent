package main

import (
	"fmt"
	"os"
	"runtime"
	"text/template"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

var (
	version   string
	revision  string
	buildDate string
	goVersion = runtime.Version()
)

var versionTmpl = template.Must(template.New("version").Parse(`
{{ .name }} (DigitalOcean Agent)

Version:     {{.version}}
Revision:    {{.revision}}
Build Date:  {{.buildDate}}
Go Version:  {{.goVersion}}
Website:     https://github.com/digitalocean/do-agent

Copyright (c) {{.year}} DigitalOcean, Inc. All rights reserved.

This work is licensed under the terms of the Apache 2.0 license.
For a copy, see <https://www.apache.org/licenses/LICENSE-2.0.html>.
`))

var buildInfo = prometheus.NewGaugeVec(
	prometheus.GaugeOpts{
		// Namespace has to be sonar or it will get filtered
		Namespace: "sonar",
		Name:      "build_info",
		Help:      "A metric with a constant '1' value labeled by version from which the agent was built.",
	},
	[]string{"version", "revision"},
).WithLabelValues(version, revision)

func init() {
	buildInfo.Set(1)
	kingpin.VersionFlag = kingpin.Flag("version", "Show the application version information").
		Short('v').
		PreAction(func(c *kingpin.ParseContext) error {
			versionTmpl.Execute(os.Stdout, map[string]string{
				"name":      "do-agent",
				"version":   version,
				"revision":  revision,
				"buildDate": buildDate,
				"goVersion": goVersion,
				"year":      fmt.Sprintf("%d", time.Now().UTC().Year()),
			})
			os.Exit(0)
			return nil
		})
	kingpin.VersionFlag.Bool()

}
