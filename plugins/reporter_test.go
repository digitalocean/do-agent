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
	"strings"
	"testing"

	"github.com/digitalocean/do-agent/metrics"
)

func TestExportReporter(t *testing.T) {
	registry := metrics.NewRegistry()
	testRef := registry.Register("test")
	registry.AddCollector(func(r metrics.Reporter) {
		r.Update(testRef, 3.1415)
	})

	// Check that the metric gets produced on the output.
	// This does not validate the format.
	reporter := NewExportReporter()
	registry.Report(reporter)

	var out bytes.Buffer
	err := reporter.Write(&out, true)
	if err != nil {
		t.Errorf("config write failure: %s", err)
	}
	t.Logf("plugin config: %q", string(out.Bytes()))
	if !strings.Contains(string(out.Bytes()), "definitions") {
		t.Errorf("config output invalid: %q", string(out.Bytes()))
	}

	out.Reset()
	err = reporter.Write(&out, false)
	if err != nil {
		t.Errorf("value write failure: %s", err)
	}
	t.Logf("plugin value: %q", string(out.Bytes()))
	if !strings.Contains(string(out.Bytes()), "3.1415") {
		t.Errorf("config value invalid: %q", string(out.Bytes()))
	}
}
