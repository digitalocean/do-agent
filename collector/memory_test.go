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

package collector

import (
	"testing"

	"github.com/digitalocean/do-agent/procfs"
)

type stubMemoryer struct {
	NewMemoryResultMemory procfs.Memory
	NewMemoryResultErr    error
}

func (s *stubMemoryer) NewMemory() (procfs.Memory, error) {
	return s.NewMemoryResultMemory, s.NewMemoryResultErr
}

// Verify that the stubMemoryer implements the procfs.Memoryer interface.
var _ procfs.Memoryer = (*stubMemoryer)(nil)

func TestRegisterMemoryMetrics(t *testing.T) {
	m := &stubMemoryer{}
	m.NewMemoryResultErr = nil
	m.NewMemoryResultMemory = procfs.Memory{}

	expectedNames := []string{
		"memory_active",
		"memory_active_anonymous",
		"memory_active_file",
		"memory_anonymous_hugepages",
		"memory_anonymous_pages",
		"memory_bounce",
		"memory_buffers",
		"memory_cached",
		"memory_commit_limit",
		"memory_committed_as",
		"memory_direct_map_1g",
		"memory_direct_map_2m",
		"memory_direct_map_4k",
		"memory_dirty",
		"memory_hardware_corrupted",
		"memory_hugepages_free",
		"memory_hugepages_reserved",
		"memory_hugepages_surplus",
		"memory_hugepages_total",
		"memory_hugepages_size",
		"memory_inactive",
		"memory_inactive_anonymous",
		"memory_inactive_file",
		"memory_kernel_stack",
		"memory_mapped",
		"memory_free",
		"memory_total",
		"memory_locked",
		"memory_nfs_unstable",
		"memory_page_tables",
		"memory_slab_reclaimable",
		"memory_slab_unreclaimable",
		"memory_shmem",
		"memory_slab",
		"memory_swap_cached",
		"memory_swap_free",
		"memory_swap_total",
		"memory_unevictable",
		"memory_virtual_malloc_chunk",
		"memory_virtual_malloc_total",
		"memory_virtual_malloc_used",
		"memory_writeback",
		"memory_writeback_temporary",
	}

	var actualNames []string

	r := &stubRegistry{}
	f := Filters{IncludeAll: true}
	RegisterMemoryMetrics(r, m.NewMemory, f)

	for i := range r.RegisterNameOpts {
		actualNames = append(actualNames, r.RegisterNameOpts[i].Name)
	}

	testForMetricNames(t, expectedNames, actualNames)

	if r.AddCollectorFunc == nil {
		t.Error("expected collector function, found none")
	}
}
