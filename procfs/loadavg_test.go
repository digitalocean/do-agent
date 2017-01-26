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

var loadTestValues = `0.04 0.03 0.05 1/84 1269`

func TestReadLoad(t *testing.T) {
	l, err := readLoad(strings.NewReader(loadTestValues))
	if err != nil {
		t.Error("Unable to read test values")
	}

	if l.Load1 != 0.04 || l.Load5 != 0.03 || l.Load15 != 0.05 {
		t.Errorf("Expected values not found in Load=%+v", l)
	}
}

func TestParseLoadValues(t *testing.T) {
	const testLine = "0.04 0.03 0.05 1/84 1269"

	l, err := parseLoad(testLine)
	if err != nil {
		t.Errorf("Unexpected error occurred while parsing \"%s\" error=%s", testLine, err)
	}

	lr := reflect.ValueOf(l)

	var loadTestValues = []struct {
		n        string
		expected float64
	}{
		{"Load1", 0.04},
		{"Load5", 0.03},
		{"Load15", 0.05},
	}

	var loadProcTestValues = []struct {
		n        string
		expected uint64
	}{
		{"RunningProcs", 1},
		{"TotalProcs", 84},
		{"LastPIDUsed", 1269},
	}

	for _, lt := range loadTestValues {
		actual := reflect.Indirect(lr).FieldByName(lt.n).Float()
		if actual != lt.expected {
			t.Errorf("Load.%s: expected %f, actual %f", lt.n, lt.expected, actual)
		}
	}

	for _, lpt := range loadProcTestValues {
		actual := reflect.Indirect(lr).FieldByName(lpt.n).Uint()
		if actual != lpt.expected {
			t.Errorf("Load.%s: expected %d, actual %d", lpt.n, lpt.expected, actual)
		}
	}
}

func TestParseLoad(t *testing.T) {
	const testFailLine = "0.04 0.03 0.05 1/84"

	_, err := parseLoad(testFailLine)
	if err == nil {
		t.Errorf("Expected error did not occur while parsing \"%s\", there aren't enough fields", testFailLine)
	}
}
