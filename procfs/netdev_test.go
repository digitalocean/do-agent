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

package procfs

import (
	"reflect"
	"strings"
	"testing"
)

const testNetworkValues = `Inter-|   Receive                                                |  Transmit
 face |bytes    packets errs drop fifo frame compressed multicast|bytes    packets errs drop fifo colls carrier compressed
    lo:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0          0
  eth0:  258319    2330    0    0    0     0          0         0   186144    1875    0    0    0     0       0          0
`

func TestNewNetwork(t *testing.T) {
	n, err := readNetwork(strings.NewReader(testNetworkValues))
	if err != nil {
		t.Errorf("Unable to read test values")
	}

	expectedLen := 2
	if len(n) != expectedLen {
		t.Errorf("Expected %d network items, actual was %d", expectedLen, len(n))
	}
}

func TestParseNetworkValues(t *testing.T) {
	const testLine = "    lo:       107       1    2    3    4     5          6         7        8       89    9    10    11     12       13          33"

	net, err := parseNetwork(testLine)
	if err != nil {
		t.Errorf("Unexpected error occured while parsing \"%s\" error=%s", testLine, err)
	}

	netr := reflect.ValueOf(net)

	var netTestValues = []struct {
		n        string
		expected uint64
	}{
		{"RXBytes", 107},
		{"RXPackets", 1},
		{"RXErrs", 2},
		{"RXDrop", 3},
		{"RXFifo", 4},
		{"RXFrame", 5},
		{"RXCompressed", 6},
		{"RXMulticast", 7},
		{"TXBytes", 8},
		{"TXPackets", 89},
		{"TXErrs", 9},
		{"TXDrop", 10},
		{"TXFifo", 11},
		{"TXColls", 12},
		{"TXCarrier", 13},
		{"TXCompressed", 33},
	}

	for _, nt := range netTestValues {
		actual := reflect.Indirect(netr).FieldByName(nt.n).Uint()
		if actual != nt.expected {
			t.Errorf("Network.%s: expected %d, actual %d", nt.n, nt.expected, actual)
		}
	}

	expectedInterface := "lo"
	if net.Interface != expectedInterface {
		t.Errorf("Network.Interface: expected %s, actual %s", expectedInterface, net.Interface)
	}
}

func TestParseNetworkFail(t *testing.T) {
	const testFailLine = "    lo:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0"

	_, err := parseNetwork(testFailLine)
	if err == nil {
		t.Errorf("Expected error did not occur while parsing \"%s\", there aren't enough fields", testFailLine)
	}
}
