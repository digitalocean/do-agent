// Copyright 2016 DigitalOcean
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"runtime"
	"time"

	"github.com/digitalocean/do-agent/bootstrap"
	"github.com/digitalocean/do-agent/collector"
	"github.com/digitalocean/do-agent/config"
	"github.com/digitalocean/do-agent/log"
	"github.com/digitalocean/do-agent/monitoringclient"
	"github.com/digitalocean/do-agent/plugins"
	"github.com/digitalocean/do-agent/procfs"
	"github.com/digitalocean/do-agent/update"

	"github.com/digitalocean/go-metadata"
	"github.com/ianschenck/envflag"
	"github.com/jpillora/backoff"
)

var (
	defaultPluginPath = "/var/lib/do-agent/plugins"

	forceUpdate        = flag.Bool("force_update", false, "Update the version of do-agent.")
	logToSyslog        = flag.Bool("log_syslog", false, "Log to syslog.")
	logLevel           = flag.String("log_level", "INFO", "Log level to log: ERROR, INFO, DEBUG")
	debugAppKey        = envflag.String("DO_AGENT_APPKEY", "", "Override AppKey")
	debugAuthToken     = envflag.String("DO_AGENT_AUTHTOKEN", "", "Override AuthToken")
	debugAuthURL       = envflag.String("DO_AGENT_AUTHENTICATION_URL", monitoringclient.AuthURL, "Override authentication URL")
	debugDropletID     = envflag.Int64("DO_AGENT_DROPLET_ID", 0, "Override Droplet ID")
	debugLocalRepoPath = envflag.String("DO_AGENT_REPO_PATH", update.RepoLocalStore, "Override Local repository path")
	debugMetadataURL   = envflag.String("DO_AGENT_METADATA_URL", monitoringclient.MetadataURL, "Override metadata URL")
	debugMetricsURL    = envflag.String("DO_AGENT_METRICS_URL", "", "Override metrics URL")
	debugUpdateURL     = envflag.String("DO_AGENT_UPDATE_URL", update.RepoURL, "Override Update URL")
	pluginPath         = envflag.String("DO_AGENT_PLUGIN_PATH", defaultPluginPath, "Override plugin path")
	procFSRoot         = envflag.String("DO_AGENT_PROCFS_ROOT", "/proc", "Override location of /proc")

	// By default, only collect these metrics _and_ any plugins metrics. In a future version of
	// the agent, the server will be requesting the metrics to gather.
	defaultMetrics = map[string]collector.Filters{
		"cpu": collector.Filters{Regexps: []*regexp.Regexp{
			regexp.MustCompile("cpu_cpu.*"),
		}},

		"disk": collector.Filters{IncludeAll: true},

		"filesystem": collector.Filters{Regexps: []*regexp.Regexp{
			regexp.MustCompile("filesystem_(free|size).*"),
		}},

		"load": collector.Filters{IncludeAll: true},

		"memory": collector.Filters{Regexps: []*regexp.Regexp{
			regexp.MustCompile("memory_(free|cached|swap*|total)"),
		}},

		// Restrict network metrics to physical nics such as 'eno1' or 'eth1'.
		// This prevents measuring VPN 'tun' or 'tap' devices and container 'veth'.
		"network": collector.Filters{Regexps: []*regexp.Regexp{
			regexp.MustCompile(`network_(receive|transmit)_(bytes|packets)_(eno|eth)\d{1,}`),
		}},

		"node": collector.Filters{IncludeAll: true},

		"process": collector.Filters{IncludeAll: true},
	}
)

func main() {
	envflag.Parse()
	flag.Parse()

	if err := log.SetLogger(*logLevel, *logToSyslog); err != nil {
		log.Fatal(err)
	}

	log.Info("Do-Agent version ", config.Version())
	log.Info("Do-Agent build ", config.Build())
	log.Info("Architecture: ", runtime.GOARCH)
	log.Info("Operating System: ", runtime.GOOS)

	if *debugMetadataURL != monitoringclient.MetadataURL {
		log.Info("Metadata URL Override: ", *debugMetadataURL)
	}
	if *debugAuthURL != monitoringclient.AuthURL {
		log.Info("Authentication URL Override: ", *debugAuthURL)
	}
	if *debugMetricsURL != "" {
		log.Info("Metrics URL Override: ", *debugMetricsURL)
	}
	if *debugAppKey != "" {
		log.Info("AppKey Override: ", *debugAppKey)
	}
	if *debugDropletID != 0 {
		log.Info("DropletID Override: ", *debugDropletID)
	}
	if *debugAuthToken != "" {
		log.Info("AuthToken Override: ", *debugAuthToken)
	}
	if *debugUpdateURL != update.RepoURL {
		log.Info("Update URL Override: ", debugUpdateURL)
	}
	if *debugLocalRepoPath != update.RepoLocalStore {
		log.Info("Local Repository Path Override: ", *debugLocalRepoPath)
	}
	if *procFSRoot != "/proc" {
		log.Info("Using alternative location for procFS: ", *procFSRoot)
	}
	if *pluginPath != defaultPluginPath {
		log.Info("Plugin path Override: ", *pluginPath)
	}
	updater := update.NewUpdate(*debugLocalRepoPath, *debugUpdateURL)

	if *forceUpdate {
		updateAgentWithExit(updater)
	}

	metadataURL, err := url.Parse(*debugMetadataURL)
	if err != nil {
		log.Fatal(err)
	}

	metadataClient := metadata.NewClient(metadata.WithBaseURL(metadataURL))
	monitoringClient := monitoringclient.NewClient(*debugAuthURL)

	errorBackoffTimer := backoff.Backoff{
		Min:    500 * time.Millisecond,
		Max:    5 * time.Minute,
		Factor: 2,
		Jitter: true,
	}

	var credentials *bootstrap.Credentials
	for {
		credentials, err = bootstrap.InitCredentials(metadataClient, monitoringClient, *debugAppKey, *debugDropletID, *debugAuthToken)
		if err == nil {
			break
		}
		log.Info("Unable to read credentials: ", err)

		if _, err = metadataClient.AuthToken(); err != nil {
			log.Fatal("do-agent requires a DigitalOcean host")
		}
		time.Sleep(errorBackoffTimer.Duration())
	}

	if credentials.AppKey == "" {
		log.Fatal("No Appkey is configured. do-agent requires a DigitalOcean host")
	}

	smc, err := monitoringclient.CreateMetricsClient(credentials.AppKey, credentials.DropletID, credentials.Region, *debugMetricsURL)
	if err != nil {
		log.Fatal("Error creating monitoring client: ", err)
	}

	updateAgentWithRestart(updater)
	lastUpdate := time.Now()
	procfs.ProcPath = *procFSRoot

	r := smc.Registry()
	collector.RegisterCPUMetrics(r, procfs.NewStat, defaultMetrics["cpu"])
	collector.RegisterDiskMetrics(r, procfs.NewDisk, defaultMetrics["disk"])
	collector.RegisterFSMetrics(r, procfs.NewMount, defaultMetrics["filesystem"])
	collector.RegisterLoadMetrics(r, procfs.NewLoad, defaultMetrics["load"])
	collector.RegisterMemoryMetrics(r, procfs.NewMemory, defaultMetrics["memory"])
	collector.RegisterNetworkMetrics(r, procfs.NewNetwork, defaultMetrics["network"])
	collector.RegisterNodeMetrics(r, procfs.NewOSRelease, defaultMetrics["node"])
	collector.RegisterProcessMetrics(r, procfs.NewProcProc, defaultMetrics["process"])
	plugins.RegisterPluginDir(r, *pluginPath)

	for {
		log.Debug("Transmitting metrics to DigitalOcean.")
		pushInterval, err := smc.SendMetrics()
		if err != nil {
			log.Error("Sending metrics to DigitalOcean: ", err)
		}

		log.Debug(fmt.Sprintf("sleeping for %d seconds", pushInterval))
		time.Sleep(time.Duration(pushInterval) * time.Second)

		if time.Now().After(lastUpdate.Add(1 * time.Hour)) {
			lastUpdate = time.Now()
			updateAgentWithRestart(updater)
		}
	}
}

// updateAgentWithRestart looks for any available updates to the agent. If an update is found, it will
// update the agent binary and reinitialize itself. If an update isn't found or fails, it will
// only log the results of its attempt.
func updateAgentWithRestart(updater update.Updater) {
	log.Info("Checking for newer version of do-agent")

	if err := updater.FetchLatestAndExec(false); err != nil {
		if err == update.ErrUpdateNotAvailable {
			log.Info("No update available")
			return
		}

		if err == update.ErrUnableToRetrieveTargets {
			log.Info("No target available for update")
			return
		}

		// covers when the agent can’t confirm that the update that is on the server is a valid
		// update because the timestamp update itself has expired.
		if _, ok := err.(update.ErrUnableToUpdateRepo); ok {
			log.Info("No repository update available")
			return
		}

		log.Errorf("Unable to update do-agent: %s\n", err)
	}
}

// updateAgentWithExit looks for any available updates to the agent. After attempting to update
// the agent it will gracefully terminate execution.
func updateAgentWithExit(updater update.Updater) {
	log.Info("Checking for newer version of do-agent")

	if err := updater.FetchLatest(true); err != nil {
		if err == update.ErrUpdateNotAvailable {
			log.Info("No update available")
			os.Exit(0)
		}

		if err == update.ErrUnableToRetrieveTargets {
			log.Info("No target available for update")
			os.Exit(0)
		}

		// covers when the agent can’t confirm that the update that is on the server is a valid
		// update because the timestamp update itself has expired.
		if _, ok := err.(update.ErrUnableToUpdateRepo); ok {
			log.Info("No repository update available")
			os.Exit(0)
		}

		log.Errorf("Unable to update do-agent: %s\n", err)
		os.Exit(1)
	}

	log.Info("Updated successfully")
	os.Exit(0)
}
