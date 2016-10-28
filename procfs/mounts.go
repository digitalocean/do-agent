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
	"strings"
)

const mountPath = "/proc/mounts"

// Mount contains the data exposed by the /proc/mounts psuedo-file
// system file.
type Mount struct {
	Device     string
	MountPoint string
	FSType     string
}

// Mounter is a collection of mount metrics exposed by the
// procfs.
type Mounter interface {
	NewMount() ([]Mount, error)
}

// NewMount collects data from the /proc/mounts system file and
// converts it into a slice of Mounts.
func NewMount() ([]Mount, error) {
	f, err := os.Open(mountPath)
	if err != nil {
		err = fmt.Errorf("Unable to collect mount metrics from %s - error: %s", mountPath, err)
		return []Mount{}, err
	}
	defer f.Close()

	return readMount(f)
}

func readMount(f io.Reader) ([]Mount, error) {
	scanner := bufio.NewScanner(f)

	var mounts []Mount

	for scanner.Scan() {
		line := scanner.Text()

		mount, err := parseMount(line)
		if err != nil {
			return []Mount{}, err
		}
		mounts = append(mounts, mount)
	}
	return mounts, nil
}

// parseMount parses a string and returns a Mount if the string is in
// the expected format.
func parseMount(line string) (Mount, error) {
	lineArray := strings.Fields(line)

	if len(lineArray) != 6 || len(lineArray) < 3 {
		err := fmt.Errorf("Unsupported %s format: %s", mountPath, line)
		return Mount{}, err
	}

	return Mount{
		Device:     lineArray[0],
		MountPoint: lineArray[1],
		FSType:     lineArray[2],
	}, nil
}
