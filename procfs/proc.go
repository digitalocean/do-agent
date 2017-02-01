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

import "github.com/prometheus/procfs"

// ProcProc contains the data exposed by various proc files in the
// pseudo-file system.
type ProcProc struct {
	PID            int
	CPUTime        float64 //value in seconds
	ResidentMemory int     //value in bytes
	VirtualMemory  int     //value in bytes
	Comm           string
	CmdLine        []string

	UserCPUTime        float64
	KernelCPUTime      float64
	ChildUserCPUTime   float64
	ChildKernelCPUTime float64
	StartTimeCPUTime   float64
}

// Procer is a collection of process metrics exposed by the
// procfs.
type Procer interface {
	NewProcProc() ([]ProcProc, error)
}

//NewProcProc collects data from various proc pseudo-file system files
//and converts it into a ProcProc structure.
func NewProcProc() ([]ProcProc, error) {
	procs, err := procfs.AllProcs()
	if err != nil {
		return []ProcProc{}, err
	}

	var ps []ProcProc

	for _, proc := range procs {
		cli, err := proc.CmdLine()
		if err != nil || len(cli) == 0 {
			continue
		}

		var p ProcProc
		p.CmdLine = cli

		p.PID = proc.PID

		comm, _ := proc.Comm()
		p.Comm = comm

		stat, err := proc.NewStat()
		if err != nil {
			continue // because the rest of the values can't be queried
		}

		p.UserCPUTime = float64(stat.UTime)
		p.KernelCPUTime = float64(stat.STime)
		p.ChildUserCPUTime = float64(stat.CUTime)
		p.ChildKernelCPUTime = float64(stat.CSTime)
		p.StartTimeCPUTime, _ = stat.StartTime()

		p.VirtualMemory = stat.VirtualMemory()
		p.ResidentMemory = stat.ResidentMemory()
		p.CPUTime = stat.CPUTime()

		ps = append(ps, p)
	}
	return ps, nil
}
