//go:build !linux

package collector

import "fmt"

func defaultStatfs(path string) (size, free float64, err error) {
	return 0, 0, fmt.Errorf("statfs not supported on this platform: %s", path)
}
