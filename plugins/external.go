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

package plugins

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/digitalocean/do-agent/log"
)

// NewExternalPluginHandler creates a new handlers for external plugin support.
func NewExternalPluginHandler(root string) *ExternalPluginHandler {
	h := &ExternalPluginHandler{
		plugins: make(map[string]*externalPlugin),
	}

	filepath.Walk(root, func(p string, f os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		mode := f.Mode()
		if f.IsDir() || !mode.IsRegular() || (mode&0111 == 0) {
			return nil
		}

		abs, err := filepath.Abs(p)
		if err != nil {
			log.Debugf("unable to get plugin path for %q: %s", p, err)
			return nil
		}

		// Found an executable file.
		h.plugins[abs] = &externalPlugin{}
		return nil
	})
	return h
}

// ExternalPluginHandler is used to manage external plugins.
// External plugins are separate programs which can be used by do-agent to
// collect metrics on its behalf.
type ExternalPluginHandler struct {
	plugins map[string]*externalPlugin
}

type externalPlugin struct {
}

// ExecResult holds the result of a plugin run.
type ExecResult struct {
	PluginPath string
	Output     []byte
	Stderr     string
	Error      error
}

// ExecuteAll runs a command on all plugins, reporting results from successful
// runs.
// Returns a mapping from plugin path to result.
func (e *ExternalPluginHandler) ExecuteAll(args ...string) []*ExecResult {
	results := make([]*ExecResult, 0, len(e.plugins))
	for binPath := range e.plugins {
		r := e.Execute(binPath, args...)
		if r.Error != nil {
			log.Errorf("unable to execute plugin %q: %s", binPath, r.Error)
			continue
		}
		results = append(results, e.Execute(binPath, args...))
	}
	return results
}

// RemovePlugin drops the plugin reference to the given path.
// This does not touch files on disk, it just removes the in-memory reference
// so that it will not be run in future calls to ExecuteAll.
func (e *ExternalPluginHandler) RemovePlugin(binPath string) {
	delete(e.plugins, binPath)
}

// Execute runs an external plugin and returns the results.
func (e *ExternalPluginHandler) Execute(binPath string, args ...string) *ExecResult {
	cmd := exec.Command(binPath)
	if len(args) > 0 {
		cmd.Args = append(cmd.Args, args...)
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	out, err := cmd.Output()

	return &ExecResult{
		PluginPath: binPath,
		Output:     out,
		Stderr:     strings.TrimSpace(string(stderrBuf.Bytes())),
		Error:      err,
	}
}
