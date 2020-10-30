package main

import (
	"strings"
	"sync"
)

const (
	ignoredMountPointFlag = "--collector.filesystem.ignored-mount-points"
	ignoredFSTypesFlag    = "--collector.filesystem.ignored-fs-types"
	ignoredMountPoints    = `^/(rootfs/)?(boot|sys|proc|dev|host|etc|var/(lib|run)/docker/[^$]+|run/docker/[^$]+)($$|/)`
)

var (
	ignoredFSTypes = strings.Join([]string{
		"aufs", "autofs", "binfmt_misc", "cifs", "cgroup", "debugfs",
		"devpts", "devtmpfs", "ecryptfs", "efivarfs", "fuse",
		"hugetlbfs", "mqueue", "nfs", "overlayfs", "proc", "pstore",
		"rpc_pipefs", "securityfs", "smb", "sysfs", "tmpfs", "tracefs",
		"squashfs", "nsfs",
	}, `|`)

	onceRegisterFilesystemFlags = new(sync.Once)
)

// registerFilesystemFlags registers filesystem cli flags.
// This should be called from within OS-specific builds since the underlying
// collectors will not be registered otherwise.
// This func can be called multiple times.
func registerFilesystemFlags() {
	onceRegisterFilesystemFlags.Do(func() {
		additionalParams = append(additionalParams, ignoredFSTypesFlag, ignoredFSTypes)
		additionalParams = append(additionalParams, ignoredMountPointFlag, ignoredMountPoints)
	})
}
