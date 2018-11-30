package compat

import (
	"fmt"
	"strings"

	dto "github.com/prometheus/client_model/go"
)

const diskSectorSize = float64(512)

// Disk converts node_exporter disk metrics from bytes to sectors
type Disk struct{}

// Name is the name of this decorator
func (d Disk) Name() string {
	return fmt.Sprintf("%T", d)
}

// Decorate converts bytes to sectors
func (Disk) Decorate(mfs []*dto.MetricFamily) {
	for _, mf := range mfs {
		n := strings.ToLower(mf.GetName())
		switch n {
		case "node_disk_read_bytes_total":
			mf.Name = sptr("sonar_disk_sectors_read")
			for _, met := range mf.GetMetric() {
				met.Counter.Value = bytesToSector(met.Counter.Value)
			}
		case "node_disk_written_bytes_total":
			mf.Name = sptr("sonar_disk_sectors_written")
			for _, met := range mf.GetMetric() {
				met.Counter.Value = bytesToSector(met.Counter.Value)
			}
		}
	}
}

func bytesToSector(val *float64) *float64 {
	v := *val
	v = v / diskSectorSize
	return &v
}
