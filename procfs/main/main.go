package main

import (
	"fmt"
	"time"

	"github.com/digitalocean/do-agent/procfs"
)

func main() {
	processes := map[int]uint64{}
	previousTotalTime := totalCPUTime()
	allProcs, err := procfs.NewProcProc()
	if err != nil {
		panic(err)
	}
	for _, p := range allProcs {
		processes[p.PID] = uint64(p.UTime + p.STime)
	}

	fmt.Printf("%-5s %-10s %6s %6s\n", "PID", "Command", "CPU %", "Time")
	for {
		totalTime := totalCPUTime()
		allProcs, err := procfs.NewProcProc()
		if err != nil {
			panic(err)
		}

		for _, p := range allProcs {
			t := uint64(p.UTime + p.STime)
			utilization := 100 * (t - processes[p.PID]) / (totalTime - previousTotalTime)
			fmt.Printf("%5d %-10s %5d%% %6d\n", p.PID, p.Comm, utilization, t)
			processes[p.PID] = t
		}
		previousTotalTime = totalTime
		time.Sleep(1 * time.Second)
	}
}

func totalCPUTime() uint64 {
	var cpu procfs.CPU

	s, err := procfs.NewStat()
	if err != nil {
		panic(err)
	}

	for _, c := range s.CPUS {
		if c.CPU == "cpu" {
			cpu = c
		}
	}

	return cpu.TotalTime()
}
