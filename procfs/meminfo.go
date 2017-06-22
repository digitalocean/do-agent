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
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/digitalocean/do-agent/log"
)

const memoryPathSuffix = "meminfo"

// Memory contains the data exposed by the /proc/meminfo pseudo-file
// system file in kb.
type Memory struct {
	MemTotal          float64 // total physical ram kb
	MemFree           float64 // unused physical ram kb
	MemAvailable      float64
	Buffers           float64 // physical ram used for buffers kb
	Cached            float64 // physical ram used as cache
	SwapCached        float64 // swap size used as cache
	Active            float64
	Inactive          float64
	ActiveAnon        float64
	InactiveAnon      float64
	ActiveFile        float64
	InactiveFile      float64
	Unevictable       float64
	Mlocked           float64
	SwapTotal         float64
	SwapFree          float64
	Dirty             float64
	Writeback         float64
	AnonPages         float64
	Mapped            float64
	Shmem             float64
	Slab              float64
	SReclaimable      float64
	SUnreclaim        float64
	KernelStack       float64
	PageTables        float64
	NFSUnstable       float64
	Bounce            float64
	WritebackTmp      float64
	CommitLimit       float64
	CommittedAS       float64
	VmallocTotal      float64
	VmallocUsed       float64
	VmallocChunk      float64
	HardwareCorrupted float64
	AnonHugePages     float64
	CmaTotal          float64
	CmaFree           float64
	HugePagesTotal    float64
	HugePagesFree     float64
	HugePagesRsvd     float64
	HugePagesSurp     float64
	Hugepagesize      float64
	DirectMap4k       float64
	DirectMap2M       float64
	DirectMap1G       float64
}

type memoryLine struct {
	field string
	value float64
}

type meminfoFieldMap map[string]*float64

// Memoryer is a collection of memory metrics exposed by the
// procfs.
type Memoryer interface {
	NewMemory() (Memory, error)
}

// Path returns the relative procfs location.
func memoryPath() string {
	return fmt.Sprintf("%s/%s", ProcPath, memoryPathSuffix)
}

// NewMemory collects data from the /proc/meminfo system file and
// converts it into a Memory structure.
func NewMemory() (Memory, error) {
	f, err := os.Open(memoryPath())
	if err != nil {
		err = fmt.Errorf("Unable to collect memory metrics from %s - error: %s", memoryPath(), err)
		return Memory{}, err
	}
	defer f.Close()

	return readMemory(f)
}

func readMemory(f io.Reader) (Memory, error) {
	scanner := bufio.NewScanner(f)
	memory := Memory{}
	memoryMap := getMeminfoFieldMap(&memory)

	for scanner.Scan() {
		line := scanner.Text()

		ml, err := parseMemory(line)
		if err != nil {
			return memory, err
		}

		if memItem, ok := memoryMap[ml.field]; ok {
			*memItem = ml.value
		} else {
			log.Debugf("meminfo field not recognized: %s", ml.field)
		}
	}
	return memory, scanner.Err()
}

// parseMemory parses single line strings from /proc/meminfo and
// creates a memoryLine struct from them. An error is created if there
// is it fails to parse the line or convert the value into a float64.
func parseMemory(line string) (memoryLine, error) {
	lineArray := strings.Fields(line)
	if len(lineArray) < 2 {
		err := fmt.Errorf("meminfo line contains less than two fields: %s", line)
		return memoryLine{}, err
	}

	ml := memoryLine{}

	ml.field = lineArray[0][0 : len(lineArray[0])-1]

	value, err := strconv.ParseFloat(lineArray[1], 64)
	if err != nil {
		err = fmt.Errorf("unable to convert meminfo value to float64: %s", line)
		return memoryLine{}, err
	}
	ml.value = value
	return ml, err
}

func getMeminfoFieldMap(memory *Memory) meminfoFieldMap {
	memoryMap := map[string]*float64{
		"MemTotal":          &memory.MemTotal,
		"MemFree":           &memory.MemFree,
		"MemAvailable":      &memory.MemAvailable,
		"Buffers":           &memory.Buffers,
		"Cached":            &memory.Cached,
		"SwapCached":        &memory.SwapCached,
		"Active":            &memory.Active,
		"Inactive":          &memory.Inactive,
		"Active(anon)":      &memory.ActiveAnon,
		"Inactive(anon)":    &memory.InactiveAnon,
		"Active(file)":      &memory.ActiveFile,
		"Inactive(file)":    &memory.InactiveFile,
		"Unevictable":       &memory.Unevictable,
		"Mlocked":           &memory.Mlocked,
		"SwapTotal":         &memory.SwapTotal,
		"SwapFree":          &memory.SwapFree,
		"Dirty":             &memory.Dirty,
		"Writeback":         &memory.Writeback,
		"AnonPages":         &memory.AnonPages,
		"Mapped":            &memory.Mapped,
		"Shmem":             &memory.Shmem,
		"Slab":              &memory.Slab,
		"SReclaimable":      &memory.SReclaimable,
		"SUnreclaim":        &memory.SUnreclaim,
		"KernelStack":       &memory.KernelStack,
		"PageTables":        &memory.PageTables,
		"NFS_Unstable":      &memory.NFSUnstable,
		"Bounce":            &memory.Bounce,
		"WritebackTmp":      &memory.WritebackTmp,
		"CommitLimit":       &memory.CommitLimit,
		"Committed_AS":      &memory.CommittedAS,
		"VmallocTotal":      &memory.VmallocTotal,
		"VmallocUsed":       &memory.VmallocUsed,
		"VmallocChunk":      &memory.VmallocChunk,
		"HardwareCorrupted": &memory.HardwareCorrupted,
		"AnonHugePages":     &memory.AnonHugePages,
		"CmaTotal":          &memory.CmaTotal,
		"CmaFree":           &memory.CmaFree,
		"HugePages_Total":   &memory.HugePagesTotal,
		"HugePages_Free":    &memory.HugePagesFree,
		"HugePages_Rsvd":    &memory.HugePagesRsvd,
		"HugePages_Surp":    &memory.HugePagesSurp,
		"Hugepagesize":      &memory.Hugepagesize,
		"DirectMap4k":       &memory.DirectMap4k,
		"DirectMap2M":       &memory.DirectMap2M,
		"DirectMap1G":       &memory.DirectMap1G,
	}
	return memoryMap
}
