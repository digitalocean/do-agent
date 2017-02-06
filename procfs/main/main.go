package main

import (
	"fmt"
	"time"

	"github.com/digitalocean/do-agent/procfs"
)

func main() {
	processes := map[int]uint64{}
	previousTotalTime := totalCPUTime()
	allProcs, _ := procfs.NewProcProc()
	for _, p := range allProcs {
		processes[p.PID] = uint64(p.UTime + p.STime)
	}

	fmt.Println("PID,Command,Utilization")
	for {
		time.Sleep(1 * time.Second)

		totalTime := totalCPUTime()
		fmt.Println(totalTime)
		allProcs, _ := procfs.NewProcProc()
		for _, p := range allProcs {
			t := uint64(p.UTime + p.STime)
			utilization := 100 * (t - processes[p.PID]) / (totalTime - previousTotalTime)
			fmt.Printf("%5d %15s %5d%% %d\n", p.PID, p.Comm, utilization, t)
			processes[p.PID] = t
		}
		previousTotalTime = totalTime
	}
}

func totalCPUTime() uint64 {
	var cpu procfs.CPU

	s, _ := procfs.NewStat()
	for _, c := range s.CPUS {
		if c.CPU == "cpu" {
			cpu = c
		}
	}

	return cpu.TotalTime()
}
