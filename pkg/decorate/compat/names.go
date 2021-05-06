package compat

import (
	"fmt"
	"strings"

	dto "github.com/prometheus/client_model/go"
)

// nameConversions is a list of metrics which differ only in name
var nameConversions = map[string]string{
	"node_network_receive_bytes_total":     "sonar_network_receive_bytes",
	"node_network_transmit_bytes_total":    "sonar_network_transmit_bytes",
	"node_memory_memtotal_bytes":           "sonar_memory_total",
	"node_memory_memfree_bytes":            "sonar_memory_free",
	"node_memory_cached_bytes":             "sonar_memory_cached",
	"node_memory_memavailable_bytes":       "sonar_memory_available",
	"node_memory_swapcached_bytes":         "sonar_memory_swap_cached",
	"node_memory_swapfree_bytes":           "sonar_memory_swap_free",
	"node_memory_swaptotal_bytes":          "sonar_memory_swap_total",
	"node_filesystem_size_bytes":           "sonar_filesystem_size",
	"node_filesystem_free_bytes":           "sonar_filesystem_free",
	"node_load1":                           "sonar_load1",
	"node_load5":                           "sonar_load5",
	"node_load15":                          "sonar_load15",
	"node_disk_reads_completed_total":      "sonar_disk_reads_completed_total",
	"node_disk_read_time_seconds_total":    "sonar_disk_read_time_seconds_total",
	"node_disk_writes_completed_total":     "sonar_disk_writes_completed_total",
	"node_disk_write_time_seconds_total":   "sonar_disk_write_time_seconds_total",
	"node_disk_discards_completed_total":   "sonar_disk_discards_completed_total",
	"node_disk_discarded_sectors_total":    "sonar_disk_discarded_sectors_total",
	"node_disk_discard_time_seconds_total": "sonar_disk_discard_time_seconds_total",
}

// Names converts node_exporter metric names to sonar names
type Names struct{}

// Name is the name of this decorator
func (n Names) Name() string {
	return fmt.Sprintf("%T", n)
}

// Decorate decorates the provided metrics for compatibility
func (Names) Decorate(mfs []*dto.MetricFamily) {
	for _, mf := range mfs {
		n := strings.ToLower(mf.GetName())
		if newName, ok := nameConversions[n]; ok {
			mf.Name = &newName
		}
	}
}

func sptr(s string) *string {
	return &s
}
