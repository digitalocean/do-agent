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

const testStatValues = `cpu  433 1 653 4451143 183 0 130 0 0 0
cpu0 185 0 345 2224578 102 0 106 0 0 0
cpu1 248 1 308 2226565 80 0 23 0 0 0
intr 339569 44 9 0 0 0 0 0 0 0 0 0 0 133 0 20059 20881 0 0 0 15068 2271 6850 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0 0
ctxt 450500
btime 1447251166
processes 1693
procs_running 1
procs_blocked 0
softirq 274487 0 86947 4456 14978 25431 0 3 83095 239 59338
`

func TestNewStat(t *testing.T) {
	s, err := readStat(strings.NewReader(testStatValues))
	if err != nil {
		t.Errorf("Unable to read test values")
		t.Error(err)
	}

	expectedLen := 3
	if len(s.CPUS) != expectedLen {
		t.Errorf("Expected %d cpus items, actual was %d", expectedLen, len(s.CPUS))
	}

	expectedInterupt := uint64(339569)
	if s.Interupt != expectedInterupt {
		t.Errorf("Expected %d cpus items, actual was %d", expectedInterupt, s.Interupt)
	}
}

func TestParseCPUValues(t *testing.T) {
	const testLine = "cpu0 185 1 345 2224578 102 2 106 3 4 5"

	c, err := parseCPU(testLine)
	if err != nil {
		t.Errorf("Unexpected error occured while parsing \"%s\" error=%s", testLine, err)
	}

	cr := reflect.ValueOf(c)

	var cpuTestValues = []struct {
		n        string
		expected uint64
	}{
		{"User", 185},
		{"Nice", 1},
		{"System", 345},
		{"Idle", 2224578},
		{"Iowait", 102},
		{"Irq", 2},
		{"Softirq", 106},
		{"Steal", 3},
		{"Guest", 4},
		{"GuestNice", 5},
	}

	for _, ct := range cpuTestValues {
		actual := reflect.Indirect(cr).FieldByName(ct.n).Uint()
		if actual != ct.expected {
			t.Errorf("CPU.%s: expected %d, actual %d", ct.n, ct.expected, actual)
		}
	}

	expectedCPU := "cpu0"
	if c.CPU != expectedCPU {
		t.Errorf("CPU.CPU: expected %s, actual %s", expectedCPU, c.CPU)
	}
}

func TestParseCPUFail(t *testing.T) {
	const testFailLine = "cpu1 248 1 308"

	_, err := parseCPU(testFailLine)
	if err == nil {
		t.Errorf("Expected error did not occur while parsing \"%s\", there aren't enough fields", testFailLine)
	}
}

func TestParseStat(t *testing.T) {
	const testLine1 = "ctxt 450500"
	const testLine2 = "procs_running 1"
	const testFailLine = "btime"

	s := Stat{}

	err := parseStat(testLine1, &s)
	if err != nil {
		t.Errorf("Unexpected error occured while parsing \"%s\" error=%s", testLine1, err)
	}

	expectedCTXT := uint64(450500)
	if s.ContextSwitch != expectedCTXT {
		t.Errorf("Expected context switches %d, actual was %d", expectedCTXT, s.ContextSwitch)
	}

	err = parseStat(testLine2, &s)
	if err != nil {
		t.Errorf("Unexpected error occured while parsing \"%s\" error=%s", testLine2, err)
	}

	expectedProcessesRunning := uint64(1)
	if s.ProcessesRunning != expectedProcessesRunning {
		t.Errorf("Expected running processes %d, actual was %d", expectedProcessesRunning, s.ProcessesRunning)
	}

	err = parseStat(testFailLine, &s)
	if err == nil {
		t.Errorf("Expected error did not occur while parsing \"%s\", there aren't enough fields", testFailLine)
	}
}
