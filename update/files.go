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
	"io"
	"io/ioutil"
	"os"
)

// Destination is a required interface for the tuf client download function
type Destination interface {
	io.Writer
	Delete() error
	Name() string
}

type tempFile struct {
	*os.File
}

// NewTempFile creates a temporary file that implements the
// Deestination interface required by go-tuf. It is a transient file
// that will be used as a temporary buffer for tuf targets.
func NewTempFile(path, prefix string) (Destination, error) {
	file, err := ioutil.TempFile(path, prefix)
	if err != nil {
		return nil, err
	}
	return &tempFile{file}, nil
}

func (t tempFile) Delete() error {
	t.File.Close()
	return os.Remove(t.Name())
}

func (t tempFile) Close() error {
	return t.Delete()
}
