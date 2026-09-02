package main

import (
	"regexp"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterFilesystemFlagsRegistersFSTypesFlag(t *testing.T) {
	// this is initialized with the _linux.go file if run on linux, but
	// this test should run on all operating systems so we initialize it
	registerFilesystemFlags()
	assert.NotEmpty(t, additionalParams)
	assert.Contains(t, additionalParams, ignoredFSTypes)
}

func TestRegisterFilesystemFlagsRegistersMountPointFlag(t *testing.T) {
	// this is initialized with the _linux.go file if run on linux, but
	// this test should run on all operating systems so we initialize it
	registerFilesystemFlags()
	assert.NotEmpty(t, additionalParams)
	assert.Contains(t, additionalParams, ignoredMountPoints)
}

func TestIgnoredMountPointsMatching(t *testing.T) {
	re := regexp.MustCompile(ignoredMountPoints)

	tests := []struct {
		mountPoint string
		ignored    bool
	}{
		{"/", false},
		{"/scratch", false},
		{"/mnt/doscratch", false},
		{"/boot", true},
		{"/proc", true},
		{"/sys", true},
		{"/dev", true},
		{"/host", true},
		{"/var/lib/docker/overlay2/abc123", true},
		{"/run/docker/netns/default", true},
		{"/run/containerd/io.containerd.runtime.v2.task/k8s.io/abc123/rootfs", true},
		{"/var/lib/containerd/io.containerd.snapshotter.v1.overlayfs/snapshots/42/fs", true},
	}

	for _, tt := range tests {
		got := re.MatchString(tt.mountPoint)
		assert.Equal(t, tt.ignored, got, "mount point %q: got ignored=%t, want ignored=%t", tt.mountPoint, got, tt.ignored)
	}
}

func TestIgnoredFSTypesMatching(t *testing.T) {
	re := regexp.MustCompile(ignoredFSTypes)

	tests := []struct {
		fsType  string
		ignored bool
	}{
		{"ext4", false},
		{"xfs", false},
		{"btrfs", false},
		{"overlay", true},
		{"overlayfs", true},
		{"tmpfs", true},
		{"sysfs", true},
		{"proc", true},
		{"nsfs", true},
	}

	for _, tt := range tests {
		got := re.MatchString(tt.fsType)
		assert.Equal(t, tt.ignored, got, "fs type %q: got ignored=%t, want ignored=%t", tt.fsType, got, tt.ignored)
	}
}
