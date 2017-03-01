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
	"testing"

	"github.com/digitalocean/do-agent/procfs"
)

type stubNetworker struct {
	NewNetworkResultNetworks []procfs.Network
	NewNetworkResultErr      error
}

func (s *stubNetworker) NewNetwork() ([]procfs.Network, error) {
	return s.NewNetworkResultNetworks, s.NewNetworkResultErr
}

// Verify that the stubNetworker implements the procfs.Networker interface.
var _ procfs.Networker = (*stubNetworker)(nil)

func TestRegisterNetworkMetrics(t *testing.T) {
	n := &stubNetworker{}
	n.NewNetworkResultErr = nil
	n.NewNetworkResultNetworks = []procfs.Network{
		procfs.Network{
			Interface:    "inet15",
			RXBytes:      uint64(1),
			RXPackets:    uint64(2),
			RXErrs:       uint64(3),
			RXDrop:       uint64(4),
			RXFifo:       uint64(5),
			RXFrame:      uint64(6),
			RXCompressed: uint64(7),
			RXMulticast:  uint64(8),
			TXBytes:      uint64(9),
			TXPackets:    uint64(10),
			TXErrs:       uint64(11),
			TXDrop:       uint64(12),
			TXFifo:       uint64(13),
			TXColls:      uint64(14),
			TXCarrier:    uint64(15),
			TXCompressed: uint64(16),
		},
	}

	expectedNames := []string{
		"network_receive_bytes",
		"network_receive_compressed",
		"network_receive_drop",
		"network_receive_errs",
		"network_receive_fifo",
		"network_receive_frame",
		"network_receive_multicast",
		"network_receive_packets",
		"network_transmit_bytes",
		"network_transmit_compressed",
		"network_transmit_drop",
		"network_transmit_errs",
		"network_transmit_fifo",
		"network_transmit_frame",
		"network_transmit_packets",
	}

	var actualNames []string

	r := &stubRegistry{}
	f := Filters{IncludeAll: true}
	RegisterNetworkMetrics(r, n.NewNetwork, f)

	for i := range r.RegisterNameOpts {
		actualNames = append(actualNames, r.RegisterNameOpts[i].Name)
	}

	testForMetricNames(t, expectedNames, actualNames)

	if r.AddCollectorFunc == nil {
		t.Error("expected collector function, found none")
	}
}
