package collector

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/log"
	"github.com/stretchr/testify/require"
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
		log.Info("metrics requested")
		io.WriteString(w, testmetrics)
	}))
	defer ts.Close()

	s, err := NewScraper("testscraper", ts.URL, 30*time.Second)
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
