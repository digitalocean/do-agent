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
	"strings"
	"syscall"

	"github.com/digitalocean/do-agent/log"
	"github.com/digitalocean/do-agent/metrics"
	"github.com/digitalocean/do-agent/procfs"
)

const (
	fsSystem = "filesystem"
)

// excludedPsudeoFSs are psudeo filesystems that are excluded from metrics.
var excludedDevices = []string{
	"autofs",
	"binfmt_misc",
	"cgroup",
	"debugfs",
	"devpts",
	"efivarfs",
	"fuse",
	"hugetlbfs",
	"lxcfs",
	"mqueue",
	"none",
	"proc",
	"pstore",
	"rootfs",
	"securityfs",
	"sysfs",
	"systemd",
	"tracefs",
	"tmpfs",
	"udev",
}

type mountFunc func() ([]procfs.Mount, error)

// isExlcuded checks if a filesystems matches the exlcudedDevice list. Regexp's were
// considered but they can be slow.
func isExcluded(c string) bool {
	for _, x := range excludedDevices {
		if strings.Contains(c, x) {
			return true
		}
	}
	return false
}

// RegisterFSMetrics registers Filesystem related metrics..
func RegisterFSMetrics(r metrics.Registry, fn mountFunc, f Filters) {
	labels := metrics.WithMeasuredLabels("device", "mountpoint", "fstype")
	available := r.Register(fsSystem+"_avail", labels)
	files := r.Register(fsSystem+"_files", labels)
	filesFree := r.Register(fsSystem+"_files_free", labels)
	free := r.Register(fsSystem+"_free", labels)
	size := r.Register(fsSystem+"_size", labels)

	r.AddCollector(func(r metrics.Reporter) {
		mounts, err := fn()
		if err != nil {
			log.Debugf("Could not gather filesystem metrics: %s", err)
			return
		}

		for _, mount := range mounts {
			if isExcluded(mount.Device) {
				log.Debugf("Ignoring mount device : %s", mount.Device)
				continue
			}

			var fsStats syscall.Statfs_t
			err := syscall.Statfs(mount.MountPoint, &fsStats)
			if err != nil {
				log.Debugf("syscall.Statfs had error on %s: %s", mount.MountPoint, err)
				continue
			}

			f.UpdateIfIncluded(r, available, float64(fsStats.Bavail)*float64(fsStats.Bsize),
				mount.Device, mount.MountPoint, mount.FSType)
			f.UpdateIfIncluded(r, files, float64(fsStats.Files),
				mount.Device, mount.MountPoint, mount.FSType)
			f.UpdateIfIncluded(r, filesFree, float64(fsStats.Ffree),
				mount.Device, mount.MountPoint, mount.FSType)
			f.UpdateIfIncluded(r, free, float64(fsStats.Bfree)*float64(fsStats.Bsize),
				mount.Device, mount.MountPoint, mount.FSType)
			f.UpdateIfIncluded(r, size, float64(fsStats.Blocks)*float64(fsStats.Bsize),
				mount.Device, mount.MountPoint, mount.FSType)
		}
	})
}
