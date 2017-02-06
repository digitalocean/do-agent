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
)

const statPath = "/proc/stat"

// CPU contains the data exposed by the /proc/stat pseudo-file system
// file for cpus.
type CPU struct {
	CPU       string
	User      uint64
	Nice      uint64
	System    uint64
	Idle      uint64
	Iowait    uint64 // since Linux 2.5.41
	Irq       uint64 // since Linux 2.6.0-test4
	Softirq   uint64 // since Linux 2.6.0-test4
	Steal     uint64 // since Linux 2.6.11
	Guest     uint64 // since Linux 2.6.24
	GuestNice uint64 // since Linux 2.6.33
}

// Stat contains the data exposed by the /proc/stat pseudo-file system
// file.
type Stat struct {
	CPUS             []CPU
	Interrupt        uint64
	ContextSwitch    uint64
	Processes        uint64
	ProcessesRunning uint64
	ProcessesBlocked uint64
}

// Stater is a collection of CPU and scheduler metrics exposed by the
// procfs.
type Stater interface {
	NewStat() (Stat, error)
}

// NewStat collects data from the /proc/stat pseudo-file system file
// and converts it into a stat struct.
func NewStat() (Stat, error) {
	f, err := os.Open(statPath)
	if err != nil {
		err = fmt.Errorf("Unable to collect stat metrics from %s - error: %s", statPath, err)
		return Stat{}, err
	}
	defer f.Close()

	return readStat(f)
}

// TotalTime (in jiffies) executed by this CPU
func (c CPU) TotalTime() uint64 {
	return c.User + c.Nice + c.System + c.Idle + c.Iowait + c.Irq + c.Softirq + c.Steal + c.Guest + c.GuestNice
}

func readStat(f io.Reader) (Stat, error) {
	var stat Stat

	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "cpu") {
			cpu, err := parseCPU(line)
			if err != nil {
				return stat, err
			}
			stat.CPUS = append(stat.CPUS, cpu)
		} else {
			err := parseStat(line, &stat)
			if err != nil {
				return stat, err
			}
		}
	}
	return stat, nil

}

// parseCPU parses a string and returns a CPU if the string is in the
// expected format.
func parseCPU(line string) (CPU, error) {
	lineArray := strings.Fields(line)

	if len(lineArray) < 5 {
		err := fmt.Errorf("Unsupported %s format: %s", statPath, line)
		return CPU{}, err
	}

	for len(lineArray) < 11 {
		lineArray = append(lineArray, "0")
	}

	user, _ := strconv.ParseUint(lineArray[1], 10, 64)
	nice, _ := strconv.ParseUint(lineArray[2], 10, 64)
	system, _ := strconv.ParseUint(lineArray[3], 10, 64)
	idle, _ := strconv.ParseUint(lineArray[4], 10, 64)
	iowait, _ := strconv.ParseUint(lineArray[5], 10, 64)
	irq, _ := strconv.ParseUint(lineArray[6], 10, 64)
	softirq, _ := strconv.ParseUint(lineArray[7], 10, 64)
	steal, _ := strconv.ParseUint(lineArray[8], 10, 64)
	guest, _ := strconv.ParseUint(lineArray[9], 10, 64)
	guestNice, _ := strconv.ParseUint(lineArray[10], 10, 64)

	metric := CPU{
		CPU:       lineArray[0],
		User:      user,
		Nice:      nice,
		System:    system,
		Idle:      idle,
		Iowait:    iowait,
		Irq:       irq,
		Softirq:   softirq,
		Steal:     steal,
		Guest:     guest,
		GuestNice: guestNice,
	}
	return metric, nil
}

// parseStat parses a string and returns a Stat if the string is in
// the expected format.
func parseStat(line string, statMetric *Stat) error {
	lineArray := strings.Fields(line)
	if len(lineArray) < 2 {
		err := fmt.Errorf("Invalid line format: \"%s\"", line)
		return err
	}

	switch lineArray[0] {
	case "intr":
		statMetric.Interrupt, _ = strconv.ParseUint(lineArray[1], 10, 64)
	case "ctxt":
		statMetric.ContextSwitch, _ = strconv.ParseUint(lineArray[1], 10, 64)
	case "processes":
		statMetric.Processes, _ = strconv.ParseUint(lineArray[1], 10, 64)
	case "procs_running":
		statMetric.ProcessesRunning, _ = strconv.ParseUint(lineArray[1], 10, 64)
	case "procs_blocked":
		statMetric.ProcessesBlocked, _ = strconv.ParseUint(lineArray[1], 10, 64)
	}
	//Default omitted due to unsupported fields.
	return nil
}
