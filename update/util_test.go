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
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"testing"
)

func TestCurrentExecPath(t *testing.T) {
	if _, err := os.Stat("/proc/self/exe"); err != nil {
		_, err2 := currentExecPath()
		if err2 == nil {
			t.Error("expected an error, got nil")
		}
	}
}

func TestParseKeys(t *testing.T) {
	const (
		key1 = ""
	)
	_, err := parseKeys(key1)
	if err == nil {
		t.Error("expected an error, got nil")
	}
}

func TestRunningTarget(t *testing.T) {
	te := fmt.Sprintf("/do-agent/do-agent_%s_%s", runtime.GOOS, runtime.GOARCH)
	ta := runningTarget()

	if ta != te {
		t.Errorf("expected %s, got %s", te, ta)
	}
}

func TestCopyFile(t *testing.T) {
	src, err := ioutil.TempFile("/tmp", "agent_test")
	if err != nil {
		t.Error("unable to create file for test.")
	}
	defer os.Remove(src.Name())

	dst, err := ioutil.TempFile("/tmp", "agent_test")
	if err != nil {
		t.Error("unable to create file for test.")
	}
	defer os.Remove(dst.Name())

	if err := ioutil.WriteFile(src.Name(), []byte("hello"), 0555); err != nil {
		t.Error("unable to write to test file.")
	}

	if err := copyFile(src.Name(), dst.Name()); err != nil {
		t.Error("copy file failed")
	}

	b, err := ioutil.ReadFile(dst.Name())
	if err != nil {
		t.Error("unable to read file contents")
	}

	if string(b) != "hello" {
		t.Error("contents written do not match contents read")
	}
}

func TestUpgradeVersion(t *testing.T) {
	var upgradeValues = []struct {
		cVersion string
		nVersion string
		fUpdate  bool
		result   bool
	}{
		{"0.0.0", "1.0.0", false, true},
		{"0.0.0", "0.1.0", false, true},
		{"0.0.0", "0.0.1", false, true},
		{"9.0.0", "0.0.0", false, false},
		{"0.9.0", "0.0.0", false, false},
		{"0.0.9", "0.0.0", false, false},
		{"1.0.0", "1.1.0", false, true},
		{"0.1.0", "0.1.1", false, true},
		{"0.0.1", "0.0.1", false, false},
		{"1.9.9", "0.10.10", false, false},
		{"1.3.2", "1.3.2", false, false},
		{"dev", "1.2.3", false, false},
		{"dev", "1.2.3", true, true},
		{"1.2.4", "dev", false, false},
		{"dev", "dev", false, false},
		{"dev", "dev", true, false},
	}

	for _, tt := range upgradeValues {
		r := upgradeVersion(tt.cVersion, tt.nVersion, tt.fUpdate)
		if r != tt.result {
			t.Errorf("version: %s new_version: %s force: %+v expected %t got %t", tt.cVersion, tt.nVersion, tt.fUpdate, tt.result, r)
		}
	}
}

func TestParseVersion(t *testing.T) {
	var eMajor, eMinor, ePatch int64 = 1, 2, 3

	v, err := parseVersion("1.2.3")
	if v.major != eMajor {
		t.Errorf("expected %d got %d", eMajor, v.major)
	}
	if v.minor != eMinor {
		t.Errorf("expected %d got %d", eMinor, v.minor)
	}
	if v.patch != ePatch {
		t.Errorf("expected %d got %d", ePatch, v.patch)
	}

	if err != nil {
		t.Error("expected nil error")
	}

	_, err = parseVersion("...")
	if err == nil {
		t.Error("expected error")
	}

}
