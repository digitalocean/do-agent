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

package update

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	tufdata "github.com/flynn/go-tuf/data"
)

const deletedTag = " (deleted)"

// currentExecPath returns the path of the current running executable
func currentExecPath() (string, error) {
	path, err := os.Readlink("/proc/self/exe")
	if err != nil {
		return "", err
	}
	path = strings.TrimSuffix(path, deletedTag)
	path = strings.TrimPrefix(path, deletedTag)
	return path, nil
}

func parseKeys(rootKeyJSON string) ([]*tufdata.Key, error) {
	var rootKeys []*tufdata.Key

	if err := json.Unmarshal([]byte(rootKeyJSON), &rootKeys); err != nil {
		return nil, ErrRootKeyParseFailed
	}
	return rootKeys, nil
}

func runningTarget() string {
	return fmt.Sprintf("/do-agent/do-agent_%s_%s", runtime.GOOS, runtime.GOARCH)
}

func copyFile(srcPath, dstPath string) error {
	buf, err := ioutil.ReadFile(srcPath)
	if err != nil {
		return ErrReadFileFailed
	}

	if err := ioutil.WriteFile(dstPath, buf, 0755); err != nil {
		if err == io.ErrShortWrite {
			return ErrWriteFileFailed
		}
		return err
	}
	return nil
}

// executeBinary calls exec() on the file located in the path. The binary
// found at that path will replace the current running process.
func executeBinary(path string) error {
	args := os.Args
	args[0] = path
	env := os.Environ()

	err := syscall.Exec(path, args, env)
	if err != nil {
		return ErrExecuteBinary
	}
	return nil
}

// Version represents a semantic version
type Version struct {
	major int64
	minor int64
	patch int64
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.major, v.minor, v.patch)
}

func upgradeVersion(current string, new string, overrideDev bool) bool {
	if overrideDev && new != "dev" {
		return true
	}

	if current == "dev" || new == "dev" {
		return false
	}

	c, err := parseVersion(current)
	if err != nil {
		return false
	}
	n, err := parseVersion(new)
	if err != nil {
		return false
	}

	if c.major < n.major {
		return true
	}
	if c.major == n.major && c.minor < n.minor {
		return true
	}
	if c.major == n.major && c.minor == n.minor && c.patch < n.patch {
		return true
	}

	return false
}

func parseVersion(version string) (*Version, error) {
	parts := strings.Split(version, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidVersionFormat
	}

	major, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return nil, ErrInvalidVersionFormat
	}
	minor, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return nil, ErrInvalidVersionFormat
	}
	patch, err := strconv.ParseInt(parts[2], 10, 64)
	if err != nil {
		return nil, ErrInvalidVersionFormat
	}

	return &Version{
		major: major,
		minor: minor,
		patch: patch,
	}, nil
}
