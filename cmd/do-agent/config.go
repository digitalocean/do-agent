package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/digitalocean/do-agent/internal/log"
	"github.com/digitalocean/do-agent/internal/process"
	"github.com/digitalocean/do-agent/pkg/clients/tsclient"
	"github.com/digitalocean/do-agent/pkg/collector"
	"github.com/digitalocean/do-agent/pkg/decorate"
	"github.com/digitalocean/do-agent/pkg/decorate/compat"
	"github.com/digitalocean/do-agent/pkg/writer"
)

var (
	config struct {
		targets       map[string]string
		metadataURL   *url.URL
		authURL       *url.URL
		sonarEndpoint string
		stdoutOnly    bool
		debug         bool
		syslog        bool
		noProcesses   bool
		noNode        bool
		kubernetes    string
		dbaas         string
	}

	// additionalParams is a list of extra command line flags to append
	// this is mostly needed for appending node_exporter flags when necessary.
	additionalParams = []string{}

	// disabledCollectors is a hash used by disableCollectors to prevent
	// duplicate entries
	disabledCollectors = map[string]interface{}{}

	kubernetesClusterUUIDUserDataPrefix = "k8saas_cluster_uuid: "
	kubernetesClusterUUIDLabel = "kubernetes_cluster_uuid"
)

const (
	defaultMetadataURL = "http://169.254.169.254/metadata"
	defaultAuthURL     = "https://sonar.digitalocean.com"
	defaultSonarURL    = ""
	defaultTimeout     = 2 * time.Second
)

func init() {
	kingpin.Flag("auth-host", "Endpoint to use for obtaining droplet app key").
		Default(defaultAuthURL).
		URLVar(&config.authURL)

	kingpin.Flag("metadata-host", "Endpoint to use for obtaining droplet metadata").
		Default(defaultMetadataURL).
		URLVar(&config.metadataURL)

	kingpin.Flag("sonar-host", "Endpoint to use for delivering metrics").
		Default(defaultSonarURL).
		StringVar(&config.sonarEndpoint)

	kingpin.Flag("stdout-only", "write all metrics to stdout only").
		BoolVar(&config.stdoutOnly)

	kingpin.Flag("debug", "display debug information to stdout").
		BoolVar(&config.debug)

	kingpin.Flag("syslog", "enable logging to syslog").
		BoolVar(&config.syslog)

	kingpin.Flag("k8s-metrics-path", "enable DO Kubernetes metrics collection (this must be a DOK8s metrics endpoint)").
		StringVar(&config.kubernetes)

	kingpin.Flag("no-collector.processes", "disable processes cpu/memory collection").Default("false").
		BoolVar(&config.noProcesses)

	kingpin.Flag("no-collector.node", "disable processes node collection").Default("false").
		BoolVar(&config.noNode)

	kingpin.Flag("dbaas-metrics-path", "enable DO DBAAS metrics collection (this must be a DO DBAAS metrics endpoint)").
		StringVar(&config.dbaas)
}

func checkConfig() error {
	var err error
	for name, uri := range config.targets {
		if _, err = url.Parse(uri); err != nil {
			return errors.Wrapf(err, "url for target %q is not valid", name)
		}
	}
	return nil
}

func initWriter() (metricWriter, throttler) {
	if config.stdoutOnly {
		return writer.NewFile(os.Stdout), &constThrottler{wait: 10 * time.Second}
	}

	tsc, err := newTimeseriesClient()
	if err != nil {
		log.Fatal("failed to connect to sonar: %+v", err)
	}
	return writer.NewSonar(tsc), tsc
}

func initDecorator() decorate.Chain {
	return decorate.Chain{
		compat.Names{},
		compat.Disk{},
		compat.CPU{},
		decorate.LowercaseNames{},
	}
}

// WrappedTSClient wraps the tsClient and adds a Name method to it
type WrappedTSClient struct {
	tsclient.Client
}

// Name returns the name of the client
func (m *WrappedTSClient) Name() string { return "tsclient" }

func newTimeseriesClient() (*WrappedTSClient, error) {
	clientOptions := []tsclient.ClientOptFn{
		tsclient.WithUserAgent(fmt.Sprintf("do-agent-%s", version)),
		tsclient.WithRadarEndpoint(config.authURL.String()),
		tsclient.WithMetadataEndpoint(config.metadataURL.String()),
	}

	if config.sonarEndpoint != "" {
		clientOptions = append(clientOptions, tsclient.WithWharfEndpoint(config.sonarEndpoint))
	}

	tsClient := tsclient.New(clientOptions...)
	wrappedTSClient := &WrappedTSClient{tsClient}

	return wrappedTSClient, nil
}

// getKubernetesClusterUUID retrieves the k8s cluster UUID from the droplet metadata
func getKubernetesClusterUUID() (string, error) {
	resp, err := http.Get(fmt.Sprintf("%s/v1/user-data", config.metadataURL))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", errors.Errorf("got status code %d while fetching kubernetes cluster UUID", resp.StatusCode)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	userData := string(body)
	return parseKubernetesClusterUUID(userData), nil
}

// parseKubernetesClusterUUID parses the user data and returns the value of the kubernetes cluster UUID
func parseKubernetesClusterUUID(userData string) string {
	userDataSplit := strings.Split(userData, "\n")
	for _, line := range userDataSplit {
		if strings.HasPrefix(line, kubernetesClusterUUIDUserDataPrefix) {
			return strings.Trim(strings.TrimPrefix(line, kubernetesClusterUUIDUserDataPrefix), "\"")
		}
	}
	return ""
}

// initCollectors initializes the prometheus collectors. By default this
// includes node_exporter and buildInfo for each remote target
func initCollectors() []prometheus.Collector {
	// buildInfo provides build information for tracking metrics internally
	cols := []prometheus.Collector{
		buildInfo,
	}

	if !config.noProcesses {
		cols = append(cols, process.NewProcessCollector())
	}

	if config.kubernetes != "" {
		kubernetesClusterUUID, err := getKubernetesClusterUUID()
		if err != nil {
			log.Error("Failed to get cluster UUID when initializing DO Kubernetes metrics: %+v", err)
		}
		var kubernetesLabels []*dto.LabelPair
		if kubernetesClusterUUID != "" {
			kubernetesLabels = append(kubernetesLabels, &dto.LabelPair{Name: &kubernetesClusterUUIDLabel, Value: &kubernetesClusterUUID})
		}
		k, err := collector.NewScraper("dokubernetes", config.kubernetes, kubernetesLabels, k8sWhitelist, defaultTimeout)
		if err != nil {
			log.Error("Failed to initialize DO Kubernetes metrics: %+v", err)
		} else {
			cols = append(cols, k)
		}
	}

	if config.dbaas != "" {
		k, err := collector.NewScraper("dodbaas", config.dbaas, nil, dbaasWhitelist, defaultTimeout)
		if err != nil {
			log.Error("Failed to initialize DO DBaaS metrics collector: %+v", err)
		} else {
			cols = append(cols, k)
		}
	}

	// create the default DO agent to collect metrics about
	// this device
	if !config.noNode {
		node, err := collector.NewNodeCollector()
		if err != nil {
			log.Fatal("failed to create DO agent: %+v", err)
		}
		log.Debug("%d node_exporter collectors were registered", len(node.Collectors()))

		for name := range node.Collectors() {
			log.Debug("node_exporter collector registered %q", name)
		}
		cols = append(cols, node)
	}

	return cols
}

// disableCollectors disables collectors by names by adding a list of
// --no-collector.<name> flags to additionalParams
func disableCollectors(names ...string) {
	f := []string{}
	for _, name := range names {
		if _, ok := disabledCollectors[name]; ok {
			// already disabled
			continue
		}

		disabledCollectors[name] = nil
		f = append(f, disableCollectorFlag(name))
	}

	additionalParams = append(additionalParams, f...)
}

// disableCollectorFlag creates the correct cli flag for the given collector name
func disableCollectorFlag(name string) string {
	return fmt.Sprintf("--no-collector.%s", name)
}
