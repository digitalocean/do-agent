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
	"github.com/digitalocean/do-agent/log"
	"github.com/digitalocean/do-agent/metrics"
	"github.com/digitalocean/do-agent/procfs"
)

const memorySystem = "memory"

var memoryMetrics = map[string]string{
	"Active":            "active",
	"Active_anon":       "active_anonymous",
	"Active_file":       "active_file",
	"AnonHugePages":     "anonymous_hugepages",
	"AnonPages":         "anonymous_pages",
	"Bounce":            "bounce",
	"Buffers":           "buffers",
	"Cached":            "cached",
	"CommitLimit":       "commit_limit",
	"Committed_AS":      "committed_as",
	"DirectMap1G":       "direct_map_1g",
	"DirectMap2M":       "direct_map_2m",
	"DirectMap4k":       "direct_map_4k",
	"Dirty":             "dirty",
	"HardwareCorrupted": "hardware_corrupted",
	"HugePages_Free":    "hugepages_free",
	"HugePages_Rsvd":    "hugepages_reserved",
	"HugePages_Surp":    "hugepages_surplus",
	"HugePages_Total":   "hugepages_total",
	"Hugepagesize":      "hugepages_size",
	"Inactive":          "inactive",
	"Inactive_anon":     "inactive_anonymous",
	"Inactive_file":     "inactive_file",
	"KernelStack":       "kernel_stack",
	"Mapped":            "mapped",
	"MemFree":           "free",
	"MemTotal":          "total",
	"Mlocked":           "locked",
	"NFS_Unstable":      "nfs_unstable",
	"PageTables":        "page_tables",
	"SReclaimable":      "slab_reclaimable",
	"SUnreclaim":        "slab_unreclaimable",
	"Shmem":             "shmem",
	"Slab":              "slab",
	"SwapCached":        "swap_cached",
	"SwapFree":          "swap_free",
	"SwapTotal":         "swap_total",
	"Unevictable":       "unevictable",
	"VmallocChunk":      "virtual_malloc_chunk",
	"VmallocTotal":      "virtual_malloc_total",
	"VmallocUsed":       "virtual_malloc_used",
	"Writeback":         "writeback",
	"WritebackTmp":      "writeback_temporary",
}

type memoryFunc func() (procfs.Memory, error)

//RegisterMemoryMetrics creates a reference to a MemoryCollector.
func RegisterMemoryMetrics(r metrics.Registry, fn memoryFunc, f Filters) {
	m := make(map[string]metrics.MetricRef)
	for procLabel, name := range memoryMetrics {
		m[procLabel] = r.Register(memorySystem + "_" + name)
	}

	r.AddCollector(func(r metrics.Reporter) {
		mem, err := fn()
		if err != nil {
			log.Debugf("couldn't get memory: %s", err)
			return
		}

		f.UpdateIfIncluded(r, m["Active"], mem.Active)
		f.UpdateIfIncluded(r, m["Active_anon"], mem.ActiveAnon)
		f.UpdateIfIncluded(r, m["Active_file"], mem.ActiveFile)
		f.UpdateIfIncluded(r, m["AnonHugePages"], mem.AnonHugePages)
		f.UpdateIfIncluded(r, m["AnonPages"], mem.AnonPages)
		f.UpdateIfIncluded(r, m["Bounce"], mem.Bounce)
		f.UpdateIfIncluded(r, m["Buffers"], mem.Buffers)
		f.UpdateIfIncluded(r, m["Cached"], mem.Cached)
		f.UpdateIfIncluded(r, m["CommitLimit"], mem.CommitLimit)
		f.UpdateIfIncluded(r, m["Committed_AS"], mem.CommittedAS)
		f.UpdateIfIncluded(r, m["DirectMap1G"], mem.DirectMap1G)
		f.UpdateIfIncluded(r, m["DirectMap2M"], mem.DirectMap2M)
		f.UpdateIfIncluded(r, m["DirectMap4k"], mem.DirectMap4k)
		f.UpdateIfIncluded(r, m["Dirty"], mem.Dirty)
		f.UpdateIfIncluded(r, m["HardwareCorrupted"], mem.HardwareCorrupted)
		f.UpdateIfIncluded(r, m["HugePages_Free"], mem.HugePagesFree)
		f.UpdateIfIncluded(r, m["HugePages_Rsvd"], mem.HugePagesRsvd)
		f.UpdateIfIncluded(r, m["HugePages_Surp"], mem.HugePagesSurp)
		f.UpdateIfIncluded(r, m["HugePages_Total"], mem.HugePagesTotal)
		f.UpdateIfIncluded(r, m["Hugepagesize"], mem.Hugepagesize)
		f.UpdateIfIncluded(r, m["Inactive"], mem.Inactive)
		f.UpdateIfIncluded(r, m["Inactive_anon"], mem.InactiveAnon)
		f.UpdateIfIncluded(r, m["Inactive_file"], mem.InactiveFile)
		f.UpdateIfIncluded(r, m["KernelStack"], mem.KernelStack)
		f.UpdateIfIncluded(r, m["Mapped"], mem.Mapped)
		f.UpdateIfIncluded(r, m["MemFree"], mem.MemFree)
		f.UpdateIfIncluded(r, m["MemTotal"], mem.MemTotal)
		f.UpdateIfIncluded(r, m["Mlocked"], mem.Mlocked)
		f.UpdateIfIncluded(r, m["NFS_Unstable"], mem.NFSUnstable)
		f.UpdateIfIncluded(r, m["PageTables"], mem.PageTables)
		f.UpdateIfIncluded(r, m["SReclaimable"], mem.SReclaimable)
		f.UpdateIfIncluded(r, m["SUnreclaim"], mem.SUnreclaim)
		f.UpdateIfIncluded(r, m["Shmem"], mem.Shmem)
		f.UpdateIfIncluded(r, m["Slab"], mem.Slab)
		f.UpdateIfIncluded(r, m["SwapCached"], mem.SwapCached)
		f.UpdateIfIncluded(r, m["SwapFree"], mem.SwapFree)
		f.UpdateIfIncluded(r, m["SwapTotal"], mem.SwapTotal)
		f.UpdateIfIncluded(r, m["Unevictable"], mem.Unevictable)
		f.UpdateIfIncluded(r, m["VmallocChunk"], mem.VmallocChunk)
		f.UpdateIfIncluded(r, m["VmallocTotal"], mem.VmallocTotal)
		f.UpdateIfIncluded(r, m["VmallocUsed"], mem.VmallocUsed)
		f.UpdateIfIncluded(r, m["Writeback"], mem.Writeback)
		f.UpdateIfIncluded(r, m["WritebackTmp"], mem.WritebackTmp)
	})
}
