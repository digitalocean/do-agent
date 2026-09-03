package collector

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/digitalocean/do-agent/internal/log"
	"github.com/prometheus/client_golang/prometheus"
)

const advancedPGMountPoint = "/pgdata"

var (
	selfFSSizeDesc = prometheus.NewDesc(
		"node_filesystem_size_bytes",
		"Filesystem size in bytes.",
		[]string{"device", "fstype", "mountpoint"},
		nil,
	)
	selfFSFreeDesc = prometheus.NewDesc(
		"node_filesystem_free_bytes",
		"Filesystem free space in bytes.",
		[]string{"device", "fstype", "mountpoint"},
		nil,
	)
)

type fsMount struct {
	device     string
	mountPoint string
	fsType     string
}

type statfsFunc func(path string) (size, free float64, err error)

// SelfFilesystemCollector emits node_filesystem_* for /pgdata from this
// process's mount table. Register it only for Advanced PG (service_type=advanced_pg).
type SelfFilesystemCollector struct {
	procfs string
	statfs statfsFunc
}

// NewSelfFilesystemCollector creates a collector for the Advanced PG data volume.
func NewSelfFilesystemCollector() *SelfFilesystemCollector {
	return &SelfFilesystemCollector{
		procfs: "/proc",
		statfs: defaultStatfs,
	}
}

// Describe implements prometheus.Collector.
func (c *SelfFilesystemCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- selfFSSizeDesc
	ch <- selfFSFreeDesc
}

// Collect implements prometheus.Collector.
func (c *SelfFilesystemCollector) Collect(ch chan<- prometheus.Metric) {
	self, err := readMounts(filepath.Join(c.procfs, "self", "mounts"))
	if err != nil {
		log.Error("self filesystem: cannot read self mounts: %v", err)
		return
	}

	m, ok := findMount(self, advancedPGMountPoint)
	if !ok {
		return
	}

	size, free, err := c.statfs(m.mountPoint)
	if err != nil {
		log.Debug("self filesystem: statfs %s: %v", m.mountPoint, err)
		return
	}
	ch <- prometheus.MustNewConstMetric(selfFSSizeDesc, prometheus.GaugeValue, size, m.device, m.fsType, m.mountPoint)
	ch <- prometheus.MustNewConstMetric(selfFSFreeDesc, prometheus.GaugeValue, free, m.device, m.fsType, m.mountPoint)
}

func readMounts(path string) ([]fsMount, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return parseMounts(f)
}

func parseMounts(r io.Reader) ([]fsMount, error) {
	var mounts []fsMount
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		parts := strings.Fields(sc.Text())
		if len(parts) < 3 {
			return nil, fmt.Errorf("malformed mount line: %q", sc.Text())
		}
		mp := strings.ReplaceAll(parts[1], `\040`, " ")
		mp = strings.ReplaceAll(mp, `\011`, "\t")
		mounts = append(mounts, fsMount{
			device:     parts[0],
			mountPoint: mp,
			fsType:     parts[2],
		})
	}
	return mounts, sc.Err()
}

func findMount(mounts []fsMount, mountPoint string) (fsMount, bool) {
	for _, m := range mounts {
		if m.mountPoint == mountPoint {
			return m, true
		}
	}
	return fsMount{}, false
}
