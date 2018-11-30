package main

import "time"

type constThrottler struct {
	wait time.Duration
}

func (c *constThrottler) WaitDuration() time.Duration {
	return c.wait
}

// Name is the name of this throttler
func (c *constThrottler) Name() string {
	return "constant"
}
