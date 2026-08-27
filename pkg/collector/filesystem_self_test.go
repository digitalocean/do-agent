package collector

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseMounts(t *testing.T) {
	in := strings.NewReader(strings.Join([]string{
		`/dev/vda1 / ext4 rw,relatime 0 0`,
		`/dev/disk/by-id/scsi-0DO_Volume_pvc-abc /pgdata ext4 ro,relatime 0 0`,
		`tmpfs /tmp tmpfs rw 0 0`,
	}, "\n") + "\n")

	got, err := parseMounts(in)
	require.NoError(t, err)
	assert.Equal(t, []fsMount{
		{device: "/dev/vda1", mountPoint: "/", fsType: "ext4"},
		{device: "/dev/disk/by-id/scsi-0DO_Volume_pvc-abc", mountPoint: "/pgdata", fsType: "ext4"},
		{device: "tmpfs", mountPoint: "/tmp", fsType: "tmpfs"},
	}, got)
}

func TestExtraMountsSkipsPID1(t *testing.T) {
	pid1 := []fsMount{{device: "/dev/vda1", mountPoint: "/", fsType: "ext4"}}
	self := []fsMount{
		{device: "/dev/vda1", mountPoint: "/", fsType: "ext4"},
		{device: "/dev/disk/by-id/scsi-0DO_Volume_pvc-abc", mountPoint: "/pgdata", fsType: "ext4"},
		{device: "tmpfs", mountPoint: "/tmp", fsType: "tmpfs"},
	}

	assert.Equal(t, []fsMount{
		{device: "/dev/disk/by-id/scsi-0DO_Volume_pvc-abc", mountPoint: "/pgdata", fsType: "ext4"},
		{device: "tmpfs", mountPoint: "/tmp", fsType: "tmpfs"},
	}, extraMounts(self, pid1))
}

func TestExtraMountsEmptyWhenSameAsPID1(t *testing.T) {
	same := []fsMount{
		{device: "/dev/vda1", mountPoint: "/", fsType: "ext4"},
		{device: "/dev/sda", mountPoint: "/mnt/volume", fsType: "ext4"},
	}
	assert.Empty(t, extraMounts(same, same))
}

func TestSelfFilesystemCollectEmitsPgdata(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, writeMounts(dir, "1", "/dev/vda1 / ext4 rw 0 0\n"))
	require.NoError(t, writeMounts(dir, "self", strings.Join([]string{
		`/dev/vda1 / ext4 rw 0 0`,
		`/dev/disk/by-id/scsi-0DO_Volume_pvc-abc /pgdata ext4 ro 0 0`,
		`tmpfs /tmp tmpfs rw 0 0`,
		`overlay /run overlay rw 0 0`,
	}, "\n")+"\n"))

	c := &SelfFilesystemCollector{
		procfs: dir,
		statfs: func(path string) (size, free float64, err error) {
			assert.Equal(t, "/pgdata", path)
			return 100, 40, nil
		},
	}

	ch := make(chan prometheus.Metric, 8)
	c.Collect(ch)
	close(ch)

	var descs []string
	for m := range ch {
		descs = append(descs, m.Desc().String())
	}
	require.Len(t, descs, 2)
	assert.Contains(t, descs[0], `fqName: "node_filesystem_size_bytes"`)
	assert.Contains(t, descs[1], `fqName: "node_filesystem_free_bytes"`)
}

func TestSelfFilesystemCollectSkipsWhenPID1Missing(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, writeMounts(dir, "self", "/dev/vda1 / ext4 rw 0 0\n"))

	c := &SelfFilesystemCollector{
		procfs: dir,
		statfs: func(string) (float64, float64, error) {
			t.Fatal("statfs should not run when pid1 mounts are missing")
			return 0, 0, nil
		},
	}

	ch := make(chan prometheus.Metric, 8)
	c.Collect(ch)
	close(ch)
	assert.Empty(t, ch)
}

func writeMounts(procfs, name, body string) error {
	dir := filepath.Join(procfs, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "mounts"), []byte(body), 0o644)
}
