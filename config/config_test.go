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

package config

import "testing"

func TestVersion(t *testing.T) {
	version = "1.2.3"
	expectedVersion := version
	if version != Version() {
		t.Errorf("version expected %s got %s", expectedVersion, Version())
	}
}

func TestVersionEmpty(t *testing.T) {
	version = ""
	expectedVersion := "dev"
	if expectedVersion != Version() {
		t.Errorf("version expected %s got %s", expectedVersion, Version())
	}
}

func TestBuild(t *testing.T) {
	build = "foo"
	expectedBuild := build
	if build != Build() {
		t.Errorf("build expected %s got %s", expectedBuild, Build())
	}
}
