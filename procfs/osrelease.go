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
)

const osReleaseSuffix = "sys/kernel/osrelease"

// OSRelease contains the data exposed by the /proc/sys/kernel/osrelease psudo-file
// system file.
type OSRelease string

// OSReleaser is a collection of os release metrics exposed by the
// procfs.
type OSReleaser interface {
	NewOSRelease() (OSRelease, error)
}

// Path returns the relative procfs location.
func osReleasePath() string {
	return fmt.Sprintf("%s/%s", ProcPath, osReleaseSuffix)
}

// NewOSRelease collects data from the /proc/sys/kernel/osrelease psudo-file system file
// and converts it into a OSRelease.
func NewOSRelease() (OSRelease, error) {
	f, err := os.Open(osReleasePath())
	if err != nil {
		err = fmt.Errorf("Unable to collect kernel version from %s - error: %s", osReleasePath(), err)
		return "", err
	}
	defer f.Close()

	return readOSRelease(f)
}

func readOSRelease(f io.Reader) (OSRelease, error) {
	scanner := bufio.NewScanner(f)

	scanner.Scan()
	release := scanner.Text()
	return OSRelease(release), nil

}
