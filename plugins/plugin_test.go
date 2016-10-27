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
	"strings"
	"testing"

	"github.com/digitalocean/do-agent/log"
	"github.com/digitalocean/do-agent/metrics"
)

func TestExternalPluginHandler(t *testing.T) {
	handler := NewExternalPluginHandler(".")
	results := handler.ExecuteAll("config")
	var found bool
	for _, res := range results {
		t.Logf("reading result from %q", res.PluginPath)
		t.Logf("-- output: %q", string(res.Output))
		if strings.Contains(string(res.Output), "definitions") {
			found = true
		}
	}
	if !found {
		t.Errorf("no plugin info found")
	}
}

func TestExternalPluginHandlerMissingDir(t *testing.T) {
	handler := NewExternalPluginHandler("/should_not_exist/i_hope")
	if len(handler.plugins) != 0 {
		t.Logf("expected 0 plugins, found: %v", handler.plugins)
	}
}

func TestExternalPluginError(t *testing.T) {
	h := NewExternalPluginHandler(".")
	var found bool
	for _, res := range h.ExecuteAll() {
		if len(res.Stderr) > 0 {
			t.Logf("found plugin error: %q", string(res.Stderr))
			if string(res.Stderr) != "intentional error" {
				t.Errorf("unexpected error: %q", res.Stderr)
			}
			found = true
		}
	}
	if !found {
		t.Errorf("expected plugin error not found")
	}
}

type mockReporter struct {
	updates []*update
}

type update struct {
	id     metrics.MetricRef
	value  float64
	labels []string
}

func (m *mockReporter) Update(id metrics.MetricRef, value float64, labels ...string) {
	m.updates = append(m.updates, &update{
		id:     id,
		value:  value,
		labels: labels,
	})
}

func TestPluginRegistry(t *testing.T) {
	registry := metrics.NewRegistry()
	RegisterPluginDir(registry, ".")

	reporter := &mockReporter{}
	registry.Report(reporter)

	var found bool
	for _, up := range reporter.updates {
		def, ok := up.id.(*metrics.Definition)
		if !ok {
			t.Errorf("unable to get metrics definition")
			continue
		}
		log.Debugf("found metric %q with value %f", def.Name, up.value)
		if def.Name == "test" {
			found = true
		}
	}
	if !found {
		t.Errorf("test plugin metric not found in result set")
	}
}
