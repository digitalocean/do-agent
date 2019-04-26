package collector

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/node_exporter/collector"
)

func fakeNodeCollector(mets []prometheus.Metric) *NodeCollector {
	return &NodeCollector{
		collectFunc: func(ch chan<- prometheus.Metric) {
			for _, m := range mets {
				ch <- m
			}
		},
		collectorsFunc: func() map[string]collector.Collector {
			return nil
		},
		describeFunc: func(ch chan<- *prometheus.Desc) {
			close(ch)
		},
	}
}

func TestNodeCollector_Collect_FiltersUnwhitelistedItems(t *testing.T) {
	metricWhitelist = []string{
		"lsadfhsadl",
		"ljfdsaiifyuoiejrw",
	}
	blacklisted := []string{
		"uyghjewkqlrh",
		"jhfdsahjkads",
		"ajslkdfhiudsa",
	}

	all := append(metricWhitelist, blacklisted...)

	mets := make([]prometheus.Metric, len(all))
	for i, item := range all {
		mets[i] = prometheus.NewCounter(prometheus.CounterOpts{
			Name: item,
		})
	}

	nc := fakeNodeCollector(mets)
	ch := make(chan prometheus.Metric, len(mets))
	nc.Collect(ch)
	close(ch)

	assert.Len(t, ch, len(metricWhitelist))
}

func TestNodeCollector_Collect_FiltersBadNetworkDevices(t *testing.T) {
	const metric = "node_network_receive_bytes_total"
	metricWhitelist = []string{metric}
	goodDevs := []string{
		"eth0",
		"eth19",
		"ens1",
		"ens4",
		"ens39",
		"eno4",
		"eno48",
	}
	badDevs := []string{
		"lo",
		"veth34uh68",
		"ajsldfjfdkasd",
	}
	allDevs := append(goodDevs, badDevs...)

	mets := make([]prometheus.Metric, len(allDevs))
	for i, device := range allDevs {
		mets[i] = prometheus.NewCounter(prometheus.CounterOpts{
			Name: metric,
			ConstLabels: prometheus.Labels{
				"device": device,
			},
			Help: "help me",
		})
	}

	nc := fakeNodeCollector(mets)
	ch := make(chan prometheus.Metric, len(mets))
	nc.Collect(ch)
	close(ch)

	assert.Len(t, ch, len(goodDevs))
	for item := range ch {
		d := item.Desc().String()
		for _, dev := range badDevs {
			assert.NotContains(t, d, fmt.Sprintf(`%q`, dev))
		}
	}
}
