package compat

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/digitalocean/do-agent/internal/log"
	dto "github.com/prometheus/client_model/go"
)

// CPU converts node_exporter cpu labels from 0-indexed to 1-indexed with prefix
type CPU struct{}

// Name is the name of this decorator
func (c CPU) Name() string {
	return fmt.Sprintf("%T", c)
}

// Decorate executes the decorator against the give metrics
func (CPU) Decorate(mfs []*dto.MetricFamily) {
	for _, mf := range mfs {
		if !strings.EqualFold(mf.GetName(), "node_cpu_seconds_total") {
			continue
		}

		mf.Name = sptr("sonar_cpu")
		for _, met := range mf.GetMetric() {
			for _, l := range met.GetLabel() {
				if !strings.EqualFold(l.GetName(), "cpu") {
					continue
				}
				num, err := strconv.Atoi(l.GetValue())
				if err != nil {
					log.Error("failed to parse cpu number: %+v", l)
					continue
				}

				l.Value = sptr(fmt.Sprintf("cpu%d", num))
			}
		}
	}
}
