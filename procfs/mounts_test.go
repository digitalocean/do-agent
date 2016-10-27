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

package procfs

import (
	"strings"
	"testing"
)

const testMountValues = `rootfs / rootfs rw 0 0
sysfs /sys sysfs rw,nosuid,nodev,noexec,relatime 0 0
proc /proc proc rw,nosuid,nodev,noexec,relatime 0 0
udev /dev devtmpfs rw,relatime,size=178072k,nr_inodes=44518,mode=755 0 0
devpts /dev/pts devpts rw,nosuid,noexec,relatime,gid=5,mode=620,ptmxmode=000 0 0
tmpfs /run tmpfs rw,nosuid,relatime,size=74852k,mode=755 0 0
/dev/mapper/precise64-root / ext4 rw,relatime,errors=remount-ro,user_xattr,barrier=1,data=ordered 0 0
none /sys/fs/fuse/connections fusectl rw,relatime 0 0
none /sys/kernel/debug debugfs rw,relatime 0 0
none /sys/kernel/security securityfs rw,relatime 0 0
none /run/lock tmpfs rw,nosuid,nodev,noexec,relatime,size=5120k 0 0
none /run/shm tmpfs rw,nosuid,nodev,relatime 0 0
/dev/sda1 /boot ext2 rw,relatime,errors=continue 0 0
rpc_pipefs /run/rpc_pipefs rpc_pipefs rw,relatime 0 0
none /vagrant vboxsf rw,nodev,relatime 0 0
`

func TestNewMount(t *testing.T) {
	m, err := readMount(strings.NewReader(testMountValues))
	if err != nil {
		t.Errorf("Unable to read test values")
	}

	// Spot checking
	if m[0].Device != "rootfs" {
		t.Errorf("device not set properly: expected=%s actual=%s", "rootfs", m[0].Device)
	}

	if m[1].MountPoint != "/sys" {
		t.Errorf("mount point not set properly: expected=%s actual=%s", "/sys", m[1].MountPoint)
	}

	if m[2].FSType != "proc" {
		t.Errorf("file system type not set properly: expected=%s actual=%s", "proc", m[2].FSType)
	}
}

func TestParseMount(t *testing.T) {
	const testLine = "/dev/yyy1 /boot ext2 rw,relatime,errors=continue 0 0"

	m, err := parseMount(testLine)
	if err != nil {
		t.Errorf("error should not be present for line=%s", testLine)
	}

	if m.Device != "/dev/yyy1" {
		t.Errorf("device not set properly: expected=%s actual=%s", "/dev/yyy1", m.Device)
	}

	if m.MountPoint != "/boot" {
		t.Errorf("mount point not set properly: expected=%s actual=%s", "/boot", m.MountPoint)
	}

	if m.FSType != "ext2" {
		t.Errorf("file system type not set properly: expected=%s actual=%s", "ext2", m.FSType)
	}
}

func TestParseMountFail(t *testing.T) {
	const testLine = "/dev/zzz1 /boot"

	_, err := parseMount(testLine)
	if err == nil {
		t.Errorf("error should be present for line=%s", testLine)
	}
}
