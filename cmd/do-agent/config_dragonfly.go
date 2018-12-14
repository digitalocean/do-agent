package main

func init() {
	registerFilesystemFlags()
	disableCollectors("boottime", "exec", "ntp", "runit", "supervisord",
		"textfile", "time",
	)
}
