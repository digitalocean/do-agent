package main

import "gopkg.in/alecthomas/kingpin.v2"

func init() {
	// Overwrite the default disk ignore list, add dm- to ignore LVM devices
	kingpin.CommandLine.GetFlag("collector.diskstats.ignored-devices").Default("^(dm-|ram|loop|fd|(h|s|v|xv)d[a-z]|nvme\\d+n\\d+p)\\d+$")

	registerFilesystemFlags()
	disableCollectors("arp", "bcache", "bonding", "buddyinfo", "conntrack",
		"drbd", "edac", "entropy", "filefd", "hwmon", "infiniband",
		"interrupts", "ipvs", "ksmd", "logind", "mdadm", "meminfo_numa",
		"mountstats", "nfs", "nfsd", "ntp", "qdisc", "runit", "sockstat",
		"supervisord", "systemd", "tcpstat", "textfile", "time", "vmstat",
		"wifi", "xfs", "zfs",
	)
}
