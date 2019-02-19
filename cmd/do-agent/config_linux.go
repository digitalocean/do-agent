package main

func init() {
	registerFilesystemFlags()
	disableCollectors("arp", "bcache", "bonding", "buddyinfo", "conntrack",
		"drbd", "edac", "entropy", "filefd", "hwmon", "infiniband",
		"interrupts", "ipvs", "ksmd", "logind", "mdadm", "meminfo_numa",
		"mountstats", "nfs", "nfsd", "ntp", "qdisc", "runit", "sockstat",
		"supervisord", "systemd", "tcpstat", "textfile", "time", "vmstat",
		"wifi", "xfs", "zfs",
	)
}
