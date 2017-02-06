package main

import "fmt"
import "github.com/digitalocean/do-agent/procfs"

func main() {
	allProcs, err := procfs.NewProcProc()
	if err != nil {
		panic(err)
	}

	fmt.Println("PID,Command,BootTime,StartTime,CPUUsage")
	for _, p := range allProcs {
		fmt.Printf("%d,%s,%.2f,%.2f,%.2f\n", p.PID, p.Comm, p.BootTime, p.StartTime, p.CPUUsage)
	}
}
