// Copyright 2016 DigitalOcean
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"regexp"

	"github.com/digitalocean/do-agent/log"
	"github.com/digitalocean/do-agent/metrics"
	"github.com/digitalocean/do-agent/procfs"
)

//from prometheus node exporter
const (
	excludedDisks = "^(ram|loop|fd|(h|s|v|xv)d[a-z])\\d+$"
	sectorSize    = 512
	diskSystem    = "disk"
)

var edp = regexp.MustCompile(excludedDisks)

type diskFunc func() ([]procfs.Disk, error)

// RegisterDiskMetrics registers disk metrics.
func RegisterDiskMetrics(r metrics.Registry, fn diskFunc, f Filters) {
	deviceLabel := metrics.WithMeasuredLabels("device")
	ioNow := r.Register(diskSystem+"_io_now", deviceLabel)
	ioTime := r.Register(diskSystem+"_io_time_ms", deviceLabel)
	ioTimeWeighted := r.Register(diskSystem+"_io_time_weighted", deviceLabel)
	readTime := r.Register(diskSystem+"_read_time_ms", deviceLabel)
	readsCompleted := r.Register(diskSystem+"_reads_completed", deviceLabel)
	readsMerged := r.Register(diskSystem+"_reads_merged", deviceLabel)
	sectorsRead := r.Register(diskSystem+"_sectors_read", deviceLabel)
	sectorsWritten := r.Register(diskSystem+"_sectors_written", deviceLabel)
	writeTime := r.Register(diskSystem+"_write_time_ms", deviceLabel)
	writesCompleted := r.Register(diskSystem+"_writes_completed", deviceLabel)
	writesMerged := r.Register(diskSystem+"_writes_merged", deviceLabel)
	bytesRead := r.Register(diskSystem+"_bytes_read", deviceLabel)
	bytesWritten := r.Register(diskSystem+"_bytes_written", deviceLabel)

	r.AddCollector(func(r metrics.Reporter) {
		disk, err := fn()
		if err != nil {
			log.Debugf("Could not gather disk metrics: %s", err)
			return
		}

		for _, value := range disk {
			if edp.MatchString(value.DeviceName) {
				log.Debugf("Excluding disk %s", value.DeviceName)
				continue
			}

			f.UpdateIfIncluded(r, ioNow, float64(value.IOInProgress), value.DeviceName)
			f.UpdateIfIncluded(r, ioTime, float64(value.TimeSpentDoingIO), value.DeviceName)
			f.UpdateIfIncluded(r, ioTimeWeighted, float64(value.WeightedTimeSpentDoingIO), value.DeviceName)
			f.UpdateIfIncluded(r, readTime, float64(value.TimeSpentReading), value.DeviceName)
			f.UpdateIfIncluded(r, readsCompleted, float64(value.ReadsCompleted), value.DeviceName)
			f.UpdateIfIncluded(r, readsMerged, float64(value.ReadsMerged), value.DeviceName)
			f.UpdateIfIncluded(r, sectorsRead, float64(value.SectorsRead), value.DeviceName)
			f.UpdateIfIncluded(r, sectorsWritten, float64(value.SectorsWritten), value.DeviceName)
			f.UpdateIfIncluded(r, writeTime, float64(value.TimeSpendWriting), value.DeviceName)
			f.UpdateIfIncluded(r, writesCompleted, float64(value.WritesCompleted), value.DeviceName)
			f.UpdateIfIncluded(r, writesMerged, float64(value.WritesMerged), value.DeviceName)
			f.UpdateIfIncluded(r, bytesRead, float64(value.ReadsMerged)*sectorSize, value.DeviceName)
			f.UpdateIfIncluded(r, bytesWritten, float64(value.WritesMerged)*sectorSize, value.DeviceName)
		}
	})
}
