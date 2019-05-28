package main

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/digitalocean/go-metadata"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/model"
	"gopkg.in/alecthomas/kingpin.v2"

	"github.com/digitalocean/do-agent/internal/flags"
	"github.com/digitalocean/do-agent/internal/log"
	"github.com/digitalocean/do-agent/pkg/clients/tsclient"
	"github.com/digitalocean/do-agent/pkg/collector"
	"github.com/digitalocean/do-agent/pkg/decorate"
	"github.com/digitalocean/do-agent/pkg/decorate/compat"
	"github.com/digitalocean/do-agent/pkg/writer"
)

var (
	config struct {
		targets          map[string]string
		metadataURL      *url.URL
		authURL          *url.URL
		sonarEndpoint    string
		stdoutOnly       bool
		debug            bool
		syslog           bool
		noProcesses      bool
		noNode           bool
		kubernetes       string
		dbaas            string
		webListenAddress string
		webListen        bool
		additionalLabels []string
	}

	// additionalParams is a list of extra command line flags to append
	// this is mostly needed for appending node_exporter flags when necessary.
	additionalParams []string

	// disabledCollectors is a hash used by disableCollectors to prevent
	// duplicate entries
	disabledCollectors = map[string]interface{}{}

	kubernetesClusterUUIDUserDataPrefix = "k8saas_cluster_uuid: "
	kubernetesClusterUUIDLabel          = "kubernetes_cluster_uuid"

	errClusterUUIDNotFound = errors.New("kubernetes cluster UUID not found")
)

const (
	defaultMetadataURL      = "http://169.254.169.254/metadata"
	defaultAuthURL          = "https://sonar.digitalocean.com"
	defaultSonarURL         = ""
	defaultTimeout          = 2 * time.Second
	defaultWebListenAddress = "127.0.0.1:9100"
)

func init() {
	kingpin.CommandLine.Name = "do-agent"

	kingpin.Flag("auth-host", "Endpoint to use for obtaining droplet app key").
		Default(defaultAuthURL).
		Envar("DO_AGENT_AUTH_URL").
		URLVar(&config.authURL)

	kingpin.Flag("metadata-host", "Endpoint to use for obtaining droplet metadata").
		Default(defaultMetadataURL).
		URLVar(&config.metadataURL)

	kingpin.Flag("sonar-host", "Endpoint to use for delivering metrics").
		Default(defaultSonarURL).
		Envar("DO_AGENT_SONAR_HOST").
		StringVar(&config.sonarEndpoint)

	kingpin.Flag("stdout-only", "write all metrics to stdout only").
		BoolVar(&config.stdoutOnly)

	kingpin.Flag("debug", "display debug information to stdout").
		BoolVar(&config.debug)

	kingpin.Flag("syslog", "enable logging to syslog").
		BoolVar(&config.syslog)

	kingpin.Flag("k8s-metrics-path", "enable DO Kubernetes metrics collection (this must be a DOK8s metrics endpoint)").
		StringVar(&config.kubernetes)

	kingpin.Flag("no-collector.processes", "disable processes cpu/memory collection").
		Default("false").
		BoolVar(&config.noProcesses)

	kingpin.Flag("no-collector.node", "disable processes node collection").
		Default("false").
		BoolVar(&config.noNode)

	kingpin.Flag("dbaas-metrics-path", "enable DO DBAAS metrics collection (this must be a DO DBAAS metrics endpoint)").
		StringVar(&config.dbaas)

	kingpin.Flag("web.listen", "enable a local endpoint for scrapeable prometheus metrics as well").
		Default("false").
		BoolVar(&config.webListen)

	kingpin.Flag("web.listen-address", `write prometheus metrics to the specified port (ex. ":9100")`).
		Default(defaultWebListenAddress).
		StringVar(&config.webListenAddress)

	kingpin.Flag("additional-label", "key value pairs for labels to add to all metrics (ex: user_id:1234)").StringsVar(&config.additionalLabels)
}

func initConfig() {
	os.Args = append(os.Args, additionalParams...)

	// read flags from cli directly first so we have access to them
	flags.Init(os.Args[1:])

	// parse all command line flags which are defined across the app
	kingpin.HelpFlag.Short('h')
	kingpin.Parse()
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

func initWriter() (metricWriter, limiter) {
	if config.stdoutOnly {
		return writer.NewFile(os.Stdout), &constThrottler{wait: 10 * time.Second}
	}

	tsc := newTimeseriesClient()
	return writer.NewSonar(tsc), tsc
}

func initDecorator() decorate.Chain {
	chain := decorate.Chain{
		compat.Names{},
		compat.Disk{},
		compat.CPU{},
		decorate.LowercaseNames{},
	}

	// If additionalLabels provided convert into decorator
	if len(config.additionalLabels) != 0 {
		chain = append(chain, decorate.LabelAppender(convertToLabelPairs(config.additionalLabels)))
	}

	return chain
}

// WrappedTSClient wraps the tsClient and adds a Name method to it
type WrappedTSClient struct {
	tsclient.Client
}

// Name returns the name of the client
func (m *WrappedTSClient) Name() string { return "tsclient" }

func newTimeseriesClient() *WrappedTSClient {
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

	return wrappedTSClient
}

// getKubernetesClusterUUID retrieves the k8s cluster UUID from the droplet metadata
func getKubernetesClusterUUID() (string, error) {
	client := metadata.NewClient(metadata.WithBaseURL(config.metadataURL))
	userData, err := client.UserData()
	if err != nil {
		return "", fmt.Errorf("failed to get user data: %+v", err)
	}
	return parseKubernetesClusterUUID(userData)
}

// parseKubernetesClusterUUID parses the user data and returns the value of the kubernetes cluster UUID
func parseKubernetesClusterUUID(userData string) (string, error) {
	userDataSplit := strings.Split(userData, "\n")
	for _, line := range userDataSplit {
		if strings.HasPrefix(line, kubernetesClusterUUIDUserDataPrefix) {
			return strings.Trim(strings.TrimPrefix(line, kubernetesClusterUUIDUserDataPrefix), "\""), nil
		}
	}
	return "", errClusterUUIDNotFound
}

// initCollectors initializes the prometheus collectors. By default this
// includes node_exporter and buildInfo for each remote target
func initCollectors() []prometheus.Collector {
	// buildInfo provides build information for tracking metrics internally
	cols := []prometheus.Collector{
		buildInfo,
		diagnosticMetric,
	}

	if config.kubernetes != "" {
		cols = appendKubernetesCollectors(cols)
	}

	if config.dbaas != "" {
		k, err := collector.NewScraper("dodbaas", config.dbaas, nil, dbaasWhitelist, collector.WithTimeout(defaultTimeout))
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

// appendKubernetesCollectors appends a kubernetes metrics collector if it can be initialized successfully
func appendKubernetesCollectors(cols []prometheus.Collector) []prometheus.Collector {
	kubernetesClusterUUID, err := getKubernetesClusterUUID()
	if err != nil {
		log.Error("Failed to get cluster UUID when initializing DO Kubernetes metrics: %+v", err)
		return cols
	}
	var kubernetesLabels []*dto.LabelPair
	kubernetesLabels = append(kubernetesLabels, &dto.LabelPair{Name: &kubernetesClusterUUIDLabel, Value: &kubernetesClusterUUID})
	k, err := collector.NewScraper("dokubernetes", config.kubernetes, kubernetesLabels, k8sWhitelist, collector.WithTimeout(defaultTimeout), collector.WithLogLevel(log.LevelDebug))
	if err != nil {
		log.Error("Failed to initialize DO Kubernetes metrics: %+v", err)
		return cols
	}
	cols = append(cols, k)
	return cols
}

// disableCollectors disables collectors by names by adding a list of
// --no-collector.<name> flags to additionalParams
func disableCollectors(names ...string) {
	f := make([]string, 0, len(names))
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

func convertToLabelPairs(s []string) []*dto.LabelPair {
	l := []*dto.LabelPair{}
	for _, lbl := range s {
		vals := strings.SplitN(lbl, ":", 2)
		if len(vals) != 2 { // require a key value pair
			log.Fatal("Bad additional-label %s, must be in the format of <key>:<value>", lbl)
		}

		if !model.LabelName(vals[0]).IsValid() {
			log.Fatal("Bad additional-label name %s", vals[0])
		}

		if !model.LabelValue(vals[1]).IsValid() {
			log.Fatal("Bad additional-label value %s", vals[1])
		}

		l = append(l, &dto.LabelPair{
			Name:  &vals[0],
			Value: &vals[1],
		})
	}

	return l
}
