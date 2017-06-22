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

const testMemoryValues = `MemTotal:         374256 kB
MemFree:          273784 kB
MemAvailable:          1 kB
Buffers:           10112 kB
Cached:            38816 kB
SwapCached:            0 kB
Active:            33920 kB
Inactive:          31516 kB
Active(anon):      16532 kB
Inactive(anon):      564 kB
Active(file):      17388 kB
Inactive(file):    30952 kB
Unevictable:           0 kB
Mlocked:               0 kB
SwapTotal:        786428 kB
SwapFree:         786428 kB
Dirty:                 0 kB
Writeback:             0 kB
AnonPages:         16480 kB
Mapped:             6652 kB
Shmem:               592 kB
Slab:              19628 kB
SReclaimable:       9360 kB
SUnreclaim:        10268 kB
KernelStack:         696 kB
PageTables:         1824 kB
NFS_Unstable:          0 kB
Bounce:                0 kB
WritebackTmp:          0 kB
CommitLimit:      973556 kB
Committed_AS:      55892 kB
VmallocTotal:   34359738367 kB
VmallocUsed:        9136 kB
VmallocChunk:   34359725884 kB
HardwareCorrupted:     0 kB
AnonHugePages:         0 kB
CmaTotal:              2 kB
CmaFree:               3 kB
HugePages_Total:       0
HugePages_Free:        0
HugePages_Rsvd:        0
HugePages_Surp:        0
Hugepagesize:       2048 kB
DirectMap4k:       57280 kB
DirectMap2M:      335872 kB
`

func TestNewMemory(t *testing.T) {
	m, err := readMemory(strings.NewReader(testMemoryValues))
	if err != nil {
		t.Errorf("Unable to read test values")
	}

	// Spot checking
	if m.MemTotal != 374256 {
		t.Errorf("Expected values not found in Memory=%+v", m)
	}

	if m.VmallocChunk != 34359725884 {
		t.Errorf("Expected values not found in Memory=%+v", m)
	}
}

func TestParseMemoryValues(t *testing.T) {
	m, err := readMemory(strings.NewReader(testMemoryValues))
	if err != nil {
		t.Errorf("Unable to open test file %s", memoryPath())
	}

	mr := reflect.ValueOf(m)

	var memoryTestValues = []struct {
		n        string
		expected float64
	}{
		{"MemTotal", 374256.0},
		{"MemFree", 273784.0},
		{"Buffers", 10112.0},
		{"Cached", 38816.0},
		{"SwapCached", 0.0},
		{"Active", 33920.0},
		{"Inactive", 31516.0},
		{"ActiveAnon", 16532.0},
		{"InactiveAnon", 564.0},
		{"ActiveFile", 17388.0},
		{"InactiveFile", 30952.0},
		{"Unevictable", 0.0},
		{"Mlocked", 0.0},
		{"SwapTotal", 786428.0},
		{"SwapFree", 786428.0},
		{"Dirty", 0.0},
		{"Writeback", 0.0},
		{"AnonPages", 16480.0},
		{"Mapped", 6652.0},
		{"Shmem", 592.0},
		{"Slab", 19628.0},
		{"SReclaimable", 9360.0},
		{"SUnreclaim", 10268.0},
		{"KernelStack", 696.0},
		{"PageTables", 1824.0},
		{"NFSUnstable", 0.0},
		{"Bounce", 0.0},
		{"WritebackTmp", 0.0},
		{"CommitLimit", 973556.0},
		{"CommittedAS", 55892.0},
		{"VmallocTotal", 34359738367.0},
		{"VmallocUsed", 9136.0},
		{"VmallocChunk", 34359725884.0},
		{"HardwareCorrupted", 0.0},
		{"AnonHugePages", 0.0},
		{"HugePagesTotal", 0.0},
		{"HugePagesFree", 0.0},
		{"HugePagesRsvd", 0.0},
		{"HugePagesSurp", 0.0},
		{"Hugepagesize", 2048.0},
		{"DirectMap4k", 57280.0},
		{"DirectMap2M", 335872.0},
		{"DirectMap1G", 0.0},
		{"MemAvailable", 1.0},
		{"CmaFree", 3.0},
		{"CmaTotal", 2.0},
	}

	for _, mt := range memoryTestValues {
		actual := reflect.Indirect(mr).FieldByName(mt.n).Float()
		if actual != mt.expected {
			t.Errorf("Memory.%s: expected %f, actual %f", mt.n, mt.expected, actual)
		}
	}
}

func TestParseMemory(t *testing.T) {
	const testLine1 = "MemTotal:         374256 kB"
	const testLine2 = "VmallocChunk:   34359725884 kB"
	const testLine3 = "HugePages_Total:       0"
	const testFailLine = "HugePages_Total:"

	ml, err := parseMemory(testLine1)
	if ml.field != "MemTotal" || ml.value != 374256 || err != nil {
		t.Errorf("Unexpected error parsing line=%s", testLine1)
	}

	ml, err = parseMemory(testLine2)
	if ml.field != "VmallocChunk" || ml.value != 34359725884 || err != nil {
		t.Errorf("Unexpected error parsing line=%s", testLine2)
	}

	ml, err = parseMemory(testLine3)
	if ml.field != "HugePages_Total" || ml.value != 0 || err != nil {
		t.Errorf("Unexpected error parsing line=%s", testLine3)
	}

	_, err = parseMemory(testFailLine)
	if err == nil {
		t.Errorf("error should be present for line=%s", testFailLine)
	}
}

func TestMeminfoFieldMap(t *testing.T) {
	memory := Memory{}
	memoryMap := getMeminfoFieldMap(&memory)

	expectedPtr := memoryMap["MemTotal"]
	if &memory.MemTotal != expectedPtr {
		t.Errorf("pointers should be equal: actual=%p expected=%p", &memory.MemTotal, expectedPtr)
	}

	expectedPtr = memoryMap["Bounce"]
	if &memory.Bounce != expectedPtr {
		t.Errorf("pointers should be equal: actual=%p expected=%p", &memory.Bounce, expectedPtr)
	}
}
