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

func TestSelfFilesystemCollectEmitsPgdata(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, writeMounts(dir, "self", strings.Join([]string{
		`/dev/vda1 / ext4 rw 0 0`,
		`/dev/disk/by-id/scsi-0DO_Volume_pvc-abc /pgdata ext4 ro 0 0`,
		`/dev/vda1 /usr ext4 ro 0 0`,
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

func TestSelfFilesystemCollectSkipsWithoutPgdata(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, writeMounts(dir, "self", strings.Join([]string{
		`/dev/vda2 / ext4 rw 0 0`,
		`/dev/vda2 /usr ext4 ro 0 0`,
	}, "\n")+"\n"))

	c := &SelfFilesystemCollector{
		procfs: dir,
		statfs: func(string) (float64, float64, error) {
			t.Fatal("statfs should not run without /pgdata")
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
