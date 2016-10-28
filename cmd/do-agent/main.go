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
	"os"
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
	debugAuthURL       = envflag.String("DO_AGENT_AUTHENTICATION_URL", monitoringclient.AuthURL, "Override authentication URL")
	debugAppKey        = envflag.String("DO_AGENT_APPKEY", "", "Override AppKey")
	debugMetricsURL    = envflag.String("DO_AGENT_METRICS_URL", "", "Override metrics URL")
	debugDropletID     = envflag.Int64("DO_AGENT_DROPLET_ID", 0, "Override Droplet ID")
	debugAuthToken     = envflag.String("DO_AGENT_AUTHTOKEN", "", "Override AuthToken")
	debugUpdateURL     = envflag.String("DO_AGENT_UPDATE_URL", update.RepoURL, "Override Update URL")
	debugLocalRepoPath = envflag.String("DO_AGENT_REPO_PATH", update.RepoLocalStore, "Override Local repository path")
	pluginPath         = envflag.String("DO_AGENT_PLUGIN_PATH", defaultPluginPath, "Override plugin path")
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
	if *pluginPath != defaultPluginPath {
		log.Info("Plugin path Override: ", *pluginPath)
	}
	updater := update.NewUpdate(*debugLocalRepoPath, *debugUpdateURL)

	if *forceUpdate {
		log.Info("Checking for newer version of do-agent")

		if err := updater.FetchLatest(true); err != nil {
			if err != update.ErrUpdateNotAvailable {
				log.Info(err)
				os.Exit(0)
			}
			log.Info("No update available")
			os.Exit(0)
		}
		log.Info("Updated successfully")
		os.Exit(0)
	}

	metadataClient := metadata.NewClient()
	monitoringClient := monitoringclient.NewClient(*debugAuthURL)

	errorBackoffTimer := backoff.Backoff{
		Min:    500 * time.Millisecond,
		Max:    5 * time.Minute,
		Factor: 2,
		Jitter: true,
	}

	var err error
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

	updateAgent(updater)
	lastUpdate := time.Now()

	r := smc.Registry()
	collector.RegisterCPUMetrics(r, procfs.NewStat)
	collector.RegisterDiskMetrics(r, procfs.NewDisk)
	collector.RegisterFSMetrics(r, procfs.NewMount)
	collector.RegisterLoadMetrics(r, procfs.NewLoad)
	collector.RegisterMemoryMetrics(r, procfs.NewMemory)
	collector.RegisterNetworkMetrics(r, procfs.NewNetwork)
	collector.RegisterNodeMetrics(r, procfs.NewOSRelease)
	collector.RegisterProcessMetrics(r, procfs.NewProcProc)
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
			updateAgent(updater)
		}
	}
}

func updateAgent(updater update.Updater) {
	log.Info("Checking for newer version of do-agent")

	if err := updater.FetchLatestAndExec(false); err != nil {
		if err == update.ErrUpdateNotAvailable {
			log.Info("No update available")
			return
		}
		log.Error(err)
		return
	}
}
