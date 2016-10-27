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

func TestCreateTufClient(t *testing.T) {
	u := &update{
		localStorePath: "/tmp",
		repositoryURL:  "not a url",
		client:         nil,
	}

	_, err := u.createTufClient()
	if err == nil {
		t.Error("expected error, recieved nil")
	}

	u2 := &update{
		localStorePath: "/tmp/aalllallalalallal",
		repositoryURL:  "http://www.digitalocean.com",
		client:         nil,
	}

	_, err2 := u2.createTufClient()
	if err2 == nil {
		t.Error("expected error, recieved nil")
	}
}
