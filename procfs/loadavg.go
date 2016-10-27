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

const loadPath = "/proc/loadavg"

// Load contains the data exposed by the /proc/loadavg psudo-file
// system file.
type Load struct {
	Load1        float64 //number of jobs in the run queue or waiting averged over 1 minute
	Load5        float64 //number of jobs in the run queue or waiting averged over 5 minutes
	Load15       float64 //number of jobs in the run queue or waiting averged over 15 minutes
	RunningProcs uint64  //Count of currently running processes
	TotalProcs   uint64  //Count of total processes
	LastPIDUsed  uint64  //Last process id used
}

// Loader is a collection of Load metrics exposed by the
// procfs.
type Loader interface {
	NewLoad() (Load, error)
}

// NewLoad collects data from the /proc/loadavg psudo-file system file
// and converts it into a Load structure.
func NewLoad() (Load, error) {
	f, err := os.Open(loadPath)
	if err != nil {
		err = fmt.Errorf("Unable to collect load metrics from %s - error: %s", loadPath, err)
		return Load{}, err
	}
	defer f.Close()

	return readLoad(f)
}

func readLoad(f io.Reader) (Load, error) {
	scanner := bufio.NewScanner(f)

	scanner.Scan()
	line := scanner.Text()

	return parseLoad(line)
}

// parseLoad parses a string and returns a Load if the string is in
// the expected format.
func parseLoad(line string) (Load, error) {
	lineArray := strings.Fields(line)

	if len(lineArray) < 5 {
		err := fmt.Errorf("Unsupported %s format: %s", loadPath, line)
		return Load{}, err
	}

	load := Load{}

	load.Load1, _ = strconv.ParseFloat(lineArray[0], 64)
	load.Load5, _ = strconv.ParseFloat(lineArray[1], 64)
	load.Load15, _ = strconv.ParseFloat(lineArray[2], 64)

	procsArray := strings.Split(lineArray[3], "/")

	load.RunningProcs, _ = strconv.ParseUint(procsArray[0], 10, 64)
	load.TotalProcs, _ = strconv.ParseUint(procsArray[1], 10, 64)
	load.LastPIDUsed, _ = strconv.ParseUint(lineArray[4], 10, 64)

	return load, nil
}
