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

import "testing"
import "time"

func TestCreateTufClient(t *testing.T) {
	u := &update{
		localStorePath: "/tmp",
		repositoryURL:  "not a url",
		interval:       3600,
		client:         nil,
	}

	_, err := u.createTufClient()
	if err == nil {
		t.Error("expected error, received nil")
	}

	u2 := &update{
		localStorePath: "/tmp/aalllallalalallal",
		repositoryURL:  "http://www.digitalocean.com",
		interval:       3600,
		client:         nil,
	}

	_, err2 := u2.createTufClient()
	if err2 == nil {
		t.Error("expected error, received nil")
	}
}

func TestInterval(t *testing.T) {
	u := &update{
		localStorePath: "/tmp",
		repositoryURL:  "not a url",
		interval:       3600,
		client:         nil,
	}

	interval1 := u.Interval()
	if interval1 != time.Second*3600 {
		t.Errorf("expected interval of 3600, received interval %s", interval1)
	}

	u2 := &update{
		localStorePath: "/tmp",
		repositoryURL:  "not a url",
		interval:       0,
		client:         nil,
	}

	interval2 := u2.Interval()
	if interval2 != 0 {
		t.Errorf("expected interval of 0, received %s", interval2)
	}

}
