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

const diskPath = "/proc/diskstats"

// Disk contains the data exposed by the /proc/diskstats pseudo-file
// system file since Linux 2.5.69.
type Disk struct {
	//major number
	MajorNumber uint64
	//minor mumber
	MinorNumber uint64
	//device name
	DeviceName string
	//reads completed successfully
	ReadsCompleted uint64
	//reads merged
	ReadsMerged uint64
	//sectors read
	SectorsRead uint64
	//time spent reading (ms)
	TimeSpentReading uint64
	//writes completed
	WritesCompleted uint64
	//writes merged
	WritesMerged uint64
	//sectors written
	SectorsWritten uint64
	//time spent writing (ms)
	TimeSpendWriting uint64
	//I/Os currently in progress
	IOInProgress uint64
	//time spent doing I/Os (ms)
	TimeSpentDoingIO uint64
	//weighted time spent doing I/Os (ms)
	WeightedTimeSpentDoingIO uint64
}

// Disker is a collection of Disk metrics exposed by the
// procfs.
type Disker interface {
	NewDisk() ([]Disk, error)
}

// NewDisk collects data from the /proc/diskstats pseudo-file system
// and converts it into an slice of Disk structures.
func NewDisk() ([]Disk, error) {
	f, err := os.Open(diskPath)
	if err != nil {
		err = fmt.Errorf("Unable to collect disk metrics from %s - error: %s", diskPath, err)
		return []Disk{}, err
	}
	defer f.Close()
	return readDisk(f)
}

func readDisk(f io.Reader) ([]Disk, error) {
	scanner := bufio.NewScanner(f)

	var disks []Disk

	for scanner.Scan() {
		line := scanner.Text()

		disk, err := parseDisk(line)
		if err != nil {
			return []Disk{}, err
		}
		disks = append(disks, disk)
	}
	return disks, scanner.Err()
}

// parseDisk parses a string and returns a Disk if the string is in
// the expected format.
func parseDisk(line string) (Disk, error) {
	lineArray := strings.Fields(line)

	if len(lineArray) < 14 {
		err := fmt.Errorf("Unsupported %s format: %s", diskPath, line)
		return Disk{}, err
	}

	disk := Disk{}

	disk.MajorNumber, _ = strconv.ParseUint(lineArray[0], 10, 64)
	disk.MinorNumber, _ = strconv.ParseUint(lineArray[1], 10, 64)
	disk.DeviceName = lineArray[2]
	disk.ReadsCompleted, _ = strconv.ParseUint(lineArray[3], 10, 64)
	disk.ReadsMerged, _ = strconv.ParseUint(lineArray[4], 10, 64)
	disk.SectorsRead, _ = strconv.ParseUint(lineArray[5], 10, 64)
	disk.TimeSpentReading, _ = strconv.ParseUint(lineArray[6], 10, 64)
	disk.WritesCompleted, _ = strconv.ParseUint(lineArray[7], 10, 64)
	disk.WritesMerged, _ = strconv.ParseUint(lineArray[8], 10, 64)
	disk.SectorsWritten, _ = strconv.ParseUint(lineArray[9], 10, 64)
	disk.TimeSpendWriting, _ = strconv.ParseUint(lineArray[10], 10, 64)
	disk.IOInProgress, _ = strconv.ParseUint(lineArray[11], 10, 64)
	disk.TimeSpentDoingIO, _ = strconv.ParseUint(lineArray[12], 10, 64)
	disk.WeightedTimeSpentDoingIO, _ = strconv.ParseUint(lineArray[13], 10, 64)

	return disk, nil
}
