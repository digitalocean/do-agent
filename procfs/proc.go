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
	"sync"

	"github.com/prometheus/procfs"
)

// ProcProc contains the data exposed by various proc files in the
// pseudo-file system.
type ProcProc struct {
	PID            int
	CPUUtilization float64 //value in % (0.0 ~ 1.0)
	ResidentMemory int     //value in bytes
	VirtualMemory  int     //value in bytes
	Comm           string
	CmdLine        []string
}

// Procer is a collection of process metrics exposed by the
// procfs.
type Procer interface {
	NewProcProc() ([]ProcProc, error)
}

var state struct {
	lock         sync.Mutex
	cpuTally     map[int]uint64
	totalCPUTime uint64
}

// NewProcProc collects data from various proc pseudo-file system files
// and converts it into a ProcProc structure.
func NewProcProc() ([]ProcProc, error) {
	allProcs, err := procfs.AllProcs()
	if err != nil {
		return []ProcProc{}, err
	}

	var (
		output          = []ProcProc{}
		newCPUTally     = map[int]uint64{}
		newTotalCPUTime = totalCPUTime()
	)

	for _, proc := range allProcs {
		cli, err := proc.CmdLine()
		if err != nil || len(cli) == 0 {
			continue
		}

		comm, err := proc.Comm()
		if err != nil {
			continue
		}

		stat, err := proc.NewStat()
		if err != nil {
			continue
		}

		var utilization float64
		newProcCPUTime := uint64(stat.UTime + stat.STime)
		newCPUTally[proc.PID] = newProcCPUTime
		if _, exists := state.cpuTally[proc.PID]; exists {
			utilization = float64(newProcCPUTime-state.cpuTally[proc.PID]) / float64(newTotalCPUTime-state.totalCPUTime)
		}

		output = append(output, ProcProc{
			CmdLine:        cli,
			PID:            proc.PID,
			Comm:           comm,
			VirtualMemory:  stat.VirtualMemory(),
			ResidentMemory: stat.ResidentMemory(),
			CPUUtilization: utilization,
		})
	}

	state.lock.Lock()
	defer state.lock.Unlock()
	state.cpuTally = newCPUTally
	state.totalCPUTime = newTotalCPUTime

	return output, nil
}

func totalCPUTime() uint64 {
	stats, err := NewStat()
	if err != nil {
		return 0
	}

	var aggregateCPU CPU
	for _, stat := range stats.CPUS {
		if stat.CPU == "cpu" {
			aggregateCPU = stat
		}
	}

	return aggregateCPU.TotalTime()
}
