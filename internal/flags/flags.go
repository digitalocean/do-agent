// Package flags reads flags passed from the command line interface.
// This package also captures parameters otherwise embedded in kingpin arguments
// further down the chain (i.e. node_exporter)
package flags

import (
	"github.com/alecthomas/kingpin/v2"
	"github.com/prometheus/procfs"
)

var (
	// ProcfsPath is the configured path to the procfs mountpoint
	ProcfsPath = procfs.DefaultMountPoint
	// SysfsPath is the configured path to the sysfs mountpoint
	SysfsPath = "/sys"
	// RootfsPath is the configured path to the rootfs mountpoint
	RootfsPath = "/"
)

// Init initializes and reads system paths from command line flags
func Init(args []string) {
	app := kingpin.New("", "")

	procfsPath := app.Flag("path.procfs", "procfs mountpoint.").Default("/proc").String()
	sysfsPath := app.Flag("path.sysfs", "sysfs mountpoint.").Default("/sys").String()
	rootfsPath := app.Flag("path.rootfs", "rootfs mountpoint.").Default("/").String()

	_, err := app.Parse(args)
	// this will always error for unknown flags passed in that aren't defined in
	// this file since we only capture the flags we're interested in. this
	// blackhole assignment silences the linter. Sue me.
	_ = err

	ProcfsPath = *procfsPath
	SysfsPath = *sysfsPath
	RootfsPath = *rootfsPath
}
