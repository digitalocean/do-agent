package collector

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/digitalocean/do-agent/internal/log"
	"github.com/prometheus/client_golang/prometheus"
)

// node_exporter reads /proc/1/mounts. In a shared-PID-namespace sidecar that
// is the pause container, which does not have volume mounts like /pgdata.
// This collector emits the same node_filesystem_* series for mounts visible
// in /proc/self/mounts that PID 1 does not have.

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

	// Keep in sync with cmd/do-agent/config_filesystem.go so we do not emit
	// tmpfs/overlay/proc junk the node collector already ignores.
	selfIgnoredMountPoints = regexp.MustCompile(`^/(rootfs/)?(boot|sys|proc|dev|host|etc|tmp|usr/(home|ports|src)|var/(audit|crash|log|mail|tmp)|var/(lib|run)/docker/[^$]+|run/docker/[^$]+)($$|/)`)
	selfIgnoredFSTypes     = regexp.MustCompile(`^(aufs|autofs|binfmt_misc|cd9660|cifs|cgroup|debugfs|devpts|devtmpfs|ecryptfs|efivarfs|fuse|hugetlbfs|mqueue|nfs|overlay|overlayfs|proc|pstore|rpc_pipefs|securityfs|smb|sysfs|tmpfs|tracefs|squashfs|nsfs)$`)
)

type fsMount struct {
	device     string
	mountPoint string
	fsType     string
}

type statfsFunc func(path string) (size, free float64, err error)

// SelfFilesystemCollector reports filesystem metrics from this process's mount
// table when they are missing from PID 1's mount table.
type SelfFilesystemCollector struct {
	procfs string
	statfs statfsFunc
}

// NewSelfFilesystemCollector creates a collector that fills in sidecar volume
// mounts node_exporter misses (for example /pgdata on Advanced PG).
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
	pid1, err := readMounts(filepath.Join(c.procfs, "1", "mounts"))
	if err != nil {
		// Missing /proc/1/mounts (hidepid) is the case node_exporter already
		// falls back from. Emitting self mounts would duplicate those series.
		log.Debug("self filesystem: skip, cannot read pid1 mounts: %v", err)
		return
	}
	self, err := readMounts(filepath.Join(c.procfs, "self", "mounts"))
	if err != nil {
		log.Error("self filesystem: cannot read self mounts: %v", err)
		return
	}

	for _, m := range extraMounts(self, pid1) {
		if selfIgnoredMountPoints.MatchString(m.mountPoint) || selfIgnoredFSTypes.MatchString(m.fsType) {
			continue
		}
		size, free, err := c.statfs(m.mountPoint)
		if err != nil {
			log.Debug("self filesystem: statfs %s: %v", m.mountPoint, err)
			continue
		}
		ch <- prometheus.MustNewConstMetric(selfFSSizeDesc, prometheus.GaugeValue, size, m.device, m.fsType, m.mountPoint)
		ch <- prometheus.MustNewConstMetric(selfFSFreeDesc, prometheus.GaugeValue, free, m.device, m.fsType, m.mountPoint)
	}
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
		mounts = append(mounts, fsMount{
			device:     parts[0],
			mountPoint: unescapeMount(parts[1]),
			fsType:     parts[2],
		})
	}
	return mounts, sc.Err()
}

func extraMounts(self, pid1 []fsMount) []fsMount {
	seenPID1 := make(map[string]struct{}, len(pid1))
	for _, m := range pid1 {
		seenPID1[m.mountPoint] = struct{}{}
	}
	seenSelf := make(map[string]struct{})
	var extra []fsMount
	for _, m := range self {
		if _, ok := seenPID1[m.mountPoint]; ok {
			continue
		}
		if _, ok := seenSelf[m.mountPoint]; ok {
			continue
		}
		seenSelf[m.mountPoint] = struct{}{}
		extra = append(extra, m)
	}
	return extra
}

func unescapeMount(p string) string {
	p = strings.ReplaceAll(p, `\040`, " ")
	p = strings.ReplaceAll(p, `\011`, "\t")
	return p
}
