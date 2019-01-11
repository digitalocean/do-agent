package process

import (
	"strconv"

	"github.com/digitalocean/do-agent/internal/flags"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/procfs"
)

type processCollector struct {
	collectFn func(chan<- prometheus.Metric)
	cpuTotal  *prometheus.Desc
	openFDs   *prometheus.Desc
	rss       *prometheus.Desc
	startTime *prometheus.Desc
}

// NewProcessCollector returns a collector which exports the current state of
// process metrics including CPU, memory and file descriptor usage as well as
// the process start time.
func NewProcessCollector() prometheus.Collector {
	c := &processCollector{
		cpuTotal: prometheus.NewDesc(
			"sonar_process_cpu_seconds_total",
			"Process user and system CPU utilization.",
			[]string{"process", "pid"}, nil,
		),
		openFDs: prometheus.NewDesc(
			"sonar_process_open_fds",
			"Number of open file descriptors.",
			[]string{"process", "pid"}, nil,
		),
		rss: prometheus.NewDesc(
			"sonar_process_resident_memory_bytes",
			"Resident memory size in bytes.",
			[]string{"process", "pid"}, nil,
		),
		startTime: prometheus.NewDesc(
			"sonar_process_start_time_seconds",
			"Start time of the process since unix epoch in seconds.",
			[]string{"process", "pid"}, nil,
		),
	}

	if _, err := procfs.NewStat(); err == nil {
		c.collectFn = c.processCollect
	} else {
		// nop
		c.collectFn = func(ch chan<- prometheus.Metric) {}
	}

	return c
}

// Describe returns all descriptions of the collector.
func (c *processCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.cpuTotal
	ch <- c.openFDs
	ch <- c.rss
	ch <- c.startTime
}

// Collect returns the current state of all metrics of the collector.
func (c *processCollector) Collect(ch chan<- prometheus.Metric) {
	c.collectFn(ch)
}

func (c *processCollector) processCollect(ch chan<- prometheus.Metric) {
	fs, err := procfs.NewFS(flags.ProcfsPath)
	if err != nil {
		return
	}

	procs, err := fs.AllProcs()
	if err != nil {
		return
	}

	for _, proc := range procs {
		stat, err := proc.NewStat()
		if err != nil {
			return
		}

		name := stat.Comm
		pid := strconv.Itoa(stat.PID)
		starttime := float64(stat.Starttime)

		ch <- prometheus.MustNewConstMetric(c.cpuTotal, prometheus.CounterValue, stat.CPUTime(), name, pid)
		ch <- prometheus.MustNewConstMetric(c.rss, prometheus.GaugeValue, float64(stat.ResidentMemory()), name, pid)
		ch <- prometheus.MustNewConstMetric(c.startTime, prometheus.GaugeValue, starttime, name, pid)

		if fds, err := proc.FileDescriptorsLen(); err == nil {
			ch <- prometheus.MustNewConstMetric(c.openFDs, prometheus.GaugeValue, float64(fds), name, pid)
		}
	}
}
