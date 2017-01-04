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

const testNetworkValuesModSpace = `Inter-| Receive | Transmit 
face |bytes packets errs drop fifo frame compressed multicast|bytes packets errs drop fifo colls carrier compressed 
lo: 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 
eth0:2345774233 3330352 0 0 0 0 0 0 1345496555 2356444 0 0 0 2 0 0
`

func TestNewNetwork(t *testing.T) {
	testCases := []struct {
		label   string
		values  string
		wantLen int
	}{
		{"baseline", testNetworkValues, 2},
		{"modSpacing", testNetworkValuesModSpace, 2},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			n, err := readNetwork(strings.NewReader(tc.values))
			if err != nil {
				t.Fatal("unable to read test values")
			}

			if tc.wantLen != len(n) {
				t.Errorf("got %d; want %d", len(n), tc.wantLen)
			}
		})
	}
}

func TestParseNetworkValues(t *testing.T) {
	const line = "    lo:       107       1    2    3    4     5          6         7        8       89    9    10    11     12       13          33"
	const lineModSpace = "eth0:1247774233 1260352 0 0 0 0 0 0 1345496819 2356759 0 0 0 0 0 0"

	type testCase struct {
		n        string
		expected uint64
	}

	testCases := []struct {
		label  string
		line   string
		netInt string
		values []testCase
	}{
		{
			"baseline",
			line,
			"lo",
			[]testCase{
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
			},
		},
		{
			"baseline",
			lineModSpace,
			"eth0",
			[]testCase{
				{"RXBytes", 1247774233},
				{"RXPackets", 1260352},
				{"RXErrs", 0},
				{"RXDrop", 0},
				{"RXFifo", 0},
				{"RXFrame", 0},
				{"RXCompressed", 0},
				{"RXMulticast", 0},
				{"TXBytes", 1345496819},
				{"TXPackets", 2356759},
				{"TXErrs", 0},
				{"TXDrop", 0},
				{"TXFifo", 0},
				{"TXColls", 0},
				{"TXCarrier", 0},
				{"TXCompressed", 0},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {

			net, err := parseNetwork(tc.line)
			if err != nil {
				t.Fatal("Unable to parse")
			}

			netr := reflect.ValueOf(net)

			for _, nt := range tc.values {
				actual := reflect.Indirect(netr).FieldByName(nt.n).Uint()
				if actual != nt.expected {
					t.Errorf("want %d; got %d", nt.expected, actual)
				}
			}

			if net.Interface != tc.netInt {
				t.Errorf("want %s; got %s", tc.netInt, net.Interface)
			}
		})
	}
}

func TestParseNetworkFail(t *testing.T) {
	const testFailLine = "    lo:       0       0    0    0    0     0          0         0        0       0    0    0    0     0       0"

	_, err := parseNetwork(testFailLine)
	if err == nil {
		t.Error("expected error did not occur")
	}
}
