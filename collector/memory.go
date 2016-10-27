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
func RegisterMemoryMetrics(r metrics.Registry, fn memoryFunc) {
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

		r.Update(m["Active"], mem.Active)
		r.Update(m["Active_anon"], mem.ActiveAnon)
		r.Update(m["Active_file"], mem.ActiveFile)
		r.Update(m["AnonHugePages"], mem.AnonHugePages)
		r.Update(m["AnonPages"], mem.AnonPages)
		r.Update(m["Bounce"], mem.Bounce)
		r.Update(m["Buffers"], mem.Buffers)
		r.Update(m["Cached"], mem.Cached)
		r.Update(m["CommitLimit"], mem.CommitLimit)
		r.Update(m["Committed_AS"], mem.CommittedAS)
		r.Update(m["DirectMap1G"], mem.DirectMap1G)
		r.Update(m["DirectMap2M"], mem.DirectMap2M)
		r.Update(m["DirectMap4k"], mem.DirectMap4k)
		r.Update(m["Dirty"], mem.Dirty)
		r.Update(m["HardwareCorrupted"], mem.HardwareCorrupted)
		r.Update(m["HugePages_Free"], mem.HugePagesFree)
		r.Update(m["HugePages_Rsvd"], mem.HugePagesRsvd)
		r.Update(m["HugePages_Surp"], mem.HugePagesSurp)
		r.Update(m["HugePages_Total"], mem.HugePagesTotal)
		r.Update(m["Hugepagesize"], mem.Hugepagesize)
		r.Update(m["Inactive"], mem.Inactive)
		r.Update(m["Inactive_anon"], mem.InactiveAnon)
		r.Update(m["Inactive_file"], mem.InactiveFile)
		r.Update(m["KernelStack"], mem.KernelStack)
		r.Update(m["Mapped"], mem.Mapped)
		r.Update(m["MemFree"], mem.MemFree)
		r.Update(m["MemTotal"], mem.MemTotal)
		r.Update(m["Mlocked"], mem.Mlocked)
		r.Update(m["NFS_Unstable"], mem.NFSUnstable)
		r.Update(m["PageTables"], mem.PageTables)
		r.Update(m["SReclaimable"], mem.SReclaimable)
		r.Update(m["SUnreclaim"], mem.SUnreclaim)
		r.Update(m["Shmem"], mem.Shmem)
		r.Update(m["Slab"], mem.Slab)
		r.Update(m["SwapCached"], mem.SwapCached)
		r.Update(m["SwapFree"], mem.SwapFree)
		r.Update(m["SwapTotal"], mem.SwapTotal)
		r.Update(m["Unevictable"], mem.Unevictable)
		r.Update(m["VmallocChunk"], mem.VmallocChunk)
		r.Update(m["VmallocTotal"], mem.VmallocTotal)
		r.Update(m["VmallocUsed"], mem.VmallocUsed)
		r.Update(m["Writeback"], mem.Writeback)
		r.Update(m["WritebackTmp"], mem.WritebackTmp)
	})
}
