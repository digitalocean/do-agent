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
	"github.com/digitalocean/do-agent/log"
	"github.com/digitalocean/do-agent/metrics"
	"github.com/digitalocean/do-agent/procfs"
)

const networkSystem = "network"

var networkNames = []string{
	"receive_bytes",
	"receive_compressed",
	"receive_drop",
	"receive_errs",
	"receive_fifo",
	"receive_frame",
	"receive_multicast",
	"receive_packets",
	"transmit_bytes",
	"transmit_compressed",
	"transmit_drop",
	"transmit_errs",
	"transmit_fifo",
	"transmit_frame",
	"transmit_packets"}

type networkFunc func() ([]procfs.Network, error)

//RegisterNetworkMetrics creates a reference to a NewtworkCollector.
func RegisterNetworkMetrics(r metrics.Registry, fn networkFunc) {
	nc := map[string]metrics.MetricRef{}
	deviceLabel := metrics.WithMeasuredLabels("device")
	for _, name := range networkNames {
		nc[name] = r.Register(networkSystem+"_"+name, deviceLabel)
	}

	r.AddCollector(func(r metrics.Reporter) {
		network, err := fn()
		if err != nil {
			log.Debugf("Could not gather network metrics: %s", err)
			return
		}

		for _, value := range network {
			r.Update(nc["receive_bytes"], float64(value.RXBytes), value.Interface)
			r.Update(nc["receive_compressed"], float64(value.RXCompressed), value.Interface)
			r.Update(nc["receive_drop"], float64(value.RXDrop), value.Interface)
			r.Update(nc["receive_errs"], float64(value.RXErrs), value.Interface)
			r.Update(nc["receive_fifo"], float64(value.RXFifo), value.Interface)
			r.Update(nc["receive_frame"], float64(value.RXFrame), value.Interface)
			r.Update(nc["receive_multicast"], float64(value.RXMulticast), value.Interface)
			r.Update(nc["receive_packets"], float64(value.RXPackets), value.Interface)
			r.Update(nc["transmit_bytes"], float64(value.TXBytes), value.Interface)
			r.Update(nc["transmit_compressed"], float64(value.TXCompressed), value.Interface)
			r.Update(nc["transmit_drop"], float64(value.TXDrop), value.Interface)
			r.Update(nc["transmit_errs"], float64(value.TXErrs), value.Interface)
			r.Update(nc["transmit_fifo"], float64(value.TXFifo), value.Interface)
			r.Update(nc["transmit_packets"], float64(value.TXPackets), value.Interface)
		}
	})
}
