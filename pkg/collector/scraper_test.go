package collector

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/require"

	"github.com/digitalocean/do-agent/internal/log"
)

var testmetrics = `# HELP kube_configmap_info Information about configmap.
# TYPE kube_configmap_info gauge
kube_configmap_info{namespace="kube-system",configmap="extension-apiserver-authentication"} 1
kube_configmap_info{namespace="kube-system",configmap="cilium-config"} 1
kube_configmap_info{namespace="kube-system",configmap="coredns"} 1
# HELP kube_configmap_created Unix creation timestamp
# TYPE kube_configmap_created gauge
kube_configmap_created{namespace="kube-system",configmap="extension-apiserver-authentication"} 1.549553255e+09
kube_configmap_created{namespace="kube-system",configmap="cilium-config"} 1.54955326e+09
kube_configmap_created{namespace="kube-system",configmap="coredns"} 1.549553261e+09
# HELP kube_configmap_metadata_resource_version Resource version representing a specific version of the configmap.
# TYPE kube_configmap_metadata_resource_version gauge
kube_configmap_metadata_resource_version{namespace="kube-system",configmap="extension-apiserver-authentication",resource_version="35"} 1
kube_configmap_metadata_resource_version{namespace="kube-system",configmap="cilium-config",resource_version="168"} 1
kube_configmap_metadata_resource_version{namespace="kube-system",configmap="coredns",resource_version="188"} 1
`

func TestScraper(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("metrics requested")
		io.WriteString(w, testmetrics)
	}))
	defer ts.Close()

	s, err := NewScraper("testscraper", ts.URL, nil, nil, 30*time.Second)
	require.NoError(t, err)

	ch := make(chan prometheus.Metric)
	go s.Collect(ch)
	for m := range ch {
		if m.Desc().String() == `Desc{fqName: "testscraper_scrape_collector_success", help: "testscraper: Whether a collector succeeded.", constLabels: {}, variableLabels: [collector]}` {
			metric := &dto.Metric{}
			m.Write(metric)
			require.Equal(t, float64(1), *metric.Gauge.Value)
			break
		}
	}
}
func TestScraperAddsKubernetesClusterUUID(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("metrics requested")
		io.WriteString(w, testmetrics)
	}))
	defer ts.Close()

	kubernetesClusterUUID := "kubernetes_cluster_uuid"
	clusterUUID := "123-345-678000"
	var kubernetesLabels []*dto.LabelPair
	kubernetesLabels = append(kubernetesLabels, &dto.LabelPair{Name: &kubernetesClusterUUID, Value: &clusterUUID})
	s, err := NewScraper("testscraper", ts.URL, kubernetesLabels, nil, 30*time.Second)
	require.NoError(t, err)

	ch := make(chan prometheus.Metric)
	go s.Collect(ch)
	for m := range ch {
		if m.Desc().String() != `Desc{fqName: "testscraper_scrape_collector_success", help: "testscraper: Whether a collector succeeded.", constLabels: {}, variableLabels: [collector]}` &&
			m.Desc().String() != `Desc{fqName: "testscraper_scrape_collector_duration_seconds", help: "testscraper: Duration of a collector scrape.", constLabels: {}, variableLabels: [collector]}` {
			metric := &dto.Metric{}
			m.Write(metric)
			require.Equal(t, float64(1), *metric.Gauge.Value)
			foundClusterUUIDLabel := false
			for _, lbl := range metric.GetLabel() {
				if lbl.GetName() == kubernetesClusterUUID && lbl.GetValue() == clusterUUID {
					foundClusterUUIDLabel = true
				}
			}
			require.True(t, foundClusterUUIDLabel)
			break
		}
	}
}



func TestWhitelist(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Debug("metrics requested")
		io.WriteString(w, testmetrics)
	}))
	defer ts.Close()

	// Only scrape kube_configmap_created
	s, err := NewScraper("testscraper", ts.URL, nil, map[string]bool{"kube_configmap_created": true}, 30*time.Second)
	require.NoError(t, err)

	ch := make(chan prometheus.Metric)
	go s.Collect(ch)
	var whitelist int
	for m := range ch {
		switch m.Desc().String() {
		case `Desc{fqName: "kube_configmap_created", help: "Unix creation timestamp", constLabels: {}, variableLabels: [namespace configmap]}`:
			whitelist++
			continue // expected whitelisted metric
		case `Desc{fqName: "testscraper_scrape_collector_success", help: "testscraper: Whether a collector succeeded.", constLabels: {}, variableLabels: [collector]}`:
			metric := &dto.Metric{}
			m.Write(metric)
			require.Equal(t, float64(1), *metric.Gauge.Value)
			return
		case `Desc{fqName: "testscraper_scrape_collector_duration_seconds", help: "testscraper: Duration of a collector scrape.", constLabels: {}, variableLabels: [collector]}`:
			continue
		default:
			t.Errorf("Unexpected metric was scraped: %v", m.Desc())
		}
	}

	// There are 3 whitelisted metrics we expected to receive
	require.Equal(t, 3, whitelist)
}
