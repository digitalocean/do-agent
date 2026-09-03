//go:build linux

package collector

import "golang.org/x/sys/unix"

func defaultStatfs(path string) (size, free float64, err error) {
	var buf unix.Statfs_t
	if err := unix.Statfs(path, &buf); err != nil {
		return 0, 0, err
	}
	bsize := float64(buf.Bsize)
	return float64(buf.Blocks) * bsize, float64(buf.Bfree) * bsize, nil
}
