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

var diskTestValues = `   1       0 ram0 0 0 0 0 0 0 0 0 0 0 0
   1       1 ram1 0 0 0 0 0 0 0 0 0 0 0
   1       2 ram2 0 0 0 0 0 0 0 0 0 0 0
   1       3 ram3 0 0 0 0 0 0 0 0 0 0 0
   1       4 ram4 0 0 0 0 0 0 0 0 0 0 0
   1       5 ram5 0 0 0 0 0 0 0 0 0 0 0
   1       6 ram6 0 0 0 0 0 0 0 0 0 0 0
   1       7 ram7 0 0 0 0 0 0 0 0 0 0 0
   1       8 ram8 0 0 0 0 0 0 0 0 0 0 0
   1       9 ram9 0 0 0 0 0 0 0 0 0 0 0
   1      10 ram10 0 0 0 0 0 0 0 0 0 0 0
   1      11 ram11 0 0 0 0 0 0 0 0 0 0 0
   1      12 ram12 0 0 0 0 0 0 0 0 0 0 0
   1      13 ram13 0 0 0 0 0 0 0 0 0 0 0
   1      14 ram14 0 0 0 0 0 0 0 0 0 0 0
   1      15 ram15 0 0 0 0 0 0 0 0 0 0 0
   7       0 loop0 0 0 0 0 0 0 0 0 0 0 0
   7       1 loop1 0 0 0 0 0 0 0 0 0 0 0
   7       2 loop2 0 0 0 0 0 0 0 0 0 0 0
   7       3 loop3 0 0 0 0 0 0 0 0 0 0 0
   7       4 loop4 0 0 0 0 0 0 0 0 0 0 0
   7       5 loop5 0 0 0 0 0 0 0 0 0 0 0
   7       6 loop6 0 0 0 0 0 0 0 0 0 0 0
   7       7 loop7 0 0 0 0 0 0 0 0 0 0 0
 253       0 vda 36472 5 1433914 9388 620084 597722 65646312 2187924 0 373096 2196720
 253       1 vda1 36283 0 1432362 9368 620084 597722 65646312 2187924 0 373080 2196696
`

func TestReadDisk(t *testing.T) {
	d, err := readDisk(strings.NewReader(diskTestValues))
	if err != nil {
		t.Error("Unable to read test values")
	}

	expectedLen := 26
	if len(d) != expectedLen {
		t.Errorf("Expected %d disk items, actual was %d", expectedLen, len(d))
	}
}

func TestParseDiskValues(t *testing.T) {
	const testLine = " 253       0 vda 36472 5 1433914 9388 620084 597722 65646312 2187924 0 373096 2196720"

	d, err := parseDisk(testLine)
	if err != nil {
		t.Errorf("Unexpected error occured while parsing \"%s\" error=%s", testLine, err)
	}

	dr := reflect.ValueOf(d)

	var diskTestValues = []struct {
		n        string
		expected uint64
	}{
		{"MajorNumber", 253},
		{"MinorNumber", 0},
		{"ReadsCompleted", 36472},
		{"ReadsMerged", 5},
		{"SectorsRead", 1433914},
		{"TimeSpentReading", 9388},
		{"WritesCompleted", 620084},
		{"WritesMerged", 597722},
		{"SectorsWritten", 65646312},
		{"TimeSpendWriting", 2187924},
		{"IOInProgress", 0},
		{"TimeSpentDoingIO", 373096},
		{"WeightedTimeSpentDoingIO", 2196720},
	}

	for _, dt := range diskTestValues {
		actual := reflect.Indirect(dr).FieldByName(dt.n).Uint()
		if actual != dt.expected {
			t.Errorf("Disk.%s:: expected %d, actual %d", dt.n, dt.expected, actual)
		}
	}

	expectedDeviceName := "vda"
	if d.DeviceName != expectedDeviceName {
		t.Errorf("Disk.DeviceName: expected %s, actual %s", expectedDeviceName, d.DeviceName)
	}
}

func TestParseDiskFail(t *testing.T) {
	const testFailLine = " 253       0 vda 36472 5 1433914 9388 620084 597722 65646312 2187924 0 373096"

	_, err := parseDisk(testFailLine)
	if err == nil {
		t.Errorf("Expected error did not occur while parsing \"%s\", there aren't enough fields", testFailLine)
	}
}
