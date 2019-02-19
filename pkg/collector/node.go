package collector

import (
	"fmt"
	"strings"

	"github.com/digitalocean/do-agent/internal/log"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/node_exporter/collector"
)

var whitelist = []string{
	"node_network_receive_bytes_total",
	"node_network_transmit_bytes_total",
	"node_memory_memtotal_bytes",
	"node_memory_memfree_bytes",
	"node_memory_cached_bytes",
	"node_memory_swapcached_bytes",
	"node_memory_swapfree_bytes",
	"node_memory_swaptotal_bytes",
	"node_filesystem_size_bytes",
	"node_filesystem_free_bytes",
	"node_disk_read_bytes_total",
	"node_disk_written_bytes_total",
	"node_cpu_seconds_total",
	"node_load1",
	"node_load5",
	"node_load15",
}

// NewNodeCollector creates a new prometheus NodeCollector
func NewNodeCollector() (*NodeCollector, error) {
	c, err := collector.NewNodeCollector()
	if err != nil {
		return nil, errors.Wrap(err, "failed to create NodeCollector")
	}

	return &NodeCollector{
		collectFunc:  c.Collect,
		describeFunc: c.Describe,
		collectorsFunc: func() map[string]collector.Collector {
			return c.Collectors
		},
	}, nil
}

// NodeCollector is a collector that collects data using
// prometheus/node_exporter. Since prometheus returns an internal type we have
// to wrap it with our own type
type NodeCollector struct {
	collectFunc    func(ch chan<- prometheus.Metric)
	describeFunc   func(ch chan<- *prometheus.Desc)
	collectorsFunc func() map[string]collector.Collector
}

// Collectors returns the list of collectors registered
func (n *NodeCollector) Collectors() map[string]collector.Collector {
	return n.collectorsFunc()
}

// Name returns the name of this collector
func (n *NodeCollector) Name() string {
	return "node"
}

// Collect collects metrics using prometheus/node_exporter
func (n *NodeCollector) Collect(ch chan<- prometheus.Metric) {
	tee := make(chan prometheus.Metric, 1)
	go func() {
		defer close(tee)
		n.collectFunc(tee)
	}()
	for m := range tee {
		// Desc doesn't allow access to underlying fields like fqName. The String() output contains
		// Desc{fqName: "node_network_transmit_bytes_total", help: "Network device statistic transmit_bytes.", constLabels: {}, variableLabels: [device]}
		// this is ugly but currently all we can do
		d := strings.ToLower(m.Desc().String())
		for _, s := range whitelist {
			if strings.Contains(d, fmt.Sprintf(`fqname: "%s"`, s)) {
				ch <- m
			} else {
				log.Debug("Node metric not whitelisted. Ignoring: %q", d)
			}
		}
	}
}

// Describe describes the metrics collected using prometheus/node_exporter
func (n *NodeCollector) Describe(ch chan<- *prometheus.Desc) {
	n.describeFunc(ch)
}
