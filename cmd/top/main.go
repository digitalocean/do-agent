package main

import (
	"fmt"
	"time"

	"github.com/digitalocean/do-agent/procfs"
)

func main() {
	fmt.Printf("%-5s %-10s %5s\n", "PID", "Command", "CPU %")
	for {
		procs, err := procfs.NewProcProc()
		if err != nil {
			panic(err)
		}

		for _, proc := range procs {
			fmt.Printf("%5d %-10s %2.2f%%\n", proc.PID, proc.Comm, 100*proc.CPUUtilization)
		}
		time.Sleep(1 * time.Second)
	}
}
