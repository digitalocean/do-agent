//go:build !nofilesystem && (linux || freebsd || openbsd || darwin || dragonfly)
// +build !nofilesystem
// +build linux freebsd openbsd darwin dragonfly

package collector

import "testing"

func stat(device, mountPoint string) filesystemStats {
	return filesystemStats{
		labels: filesystemLabels{
			device:     device,
			mountPoint: mountPoint,
			fsType:     "ext4",
		},
	}
}

// TestDedupeByDevice_CageFSBindMountFanOut reproduces the exact shape of
// the real bug report: one device (/dev/vda1) bind-mounted by CageFS into
// dozens of skeleton paths, alongside its one real mountpoint ("/"). All
// of them should collapse down to the single "/" entry.
func TestDedupeByDevice_CageFSBindMountFanOut(t *testing.T) {
	in := []filesystemStats{
		stat("/dev/vda1", "/usr/share/cagefs-skeleton/usr/local/cpanel/base"),
		stat("/dev/vda1", "/usr/share/cagefs-skeleton/opt/cpanel/ea-php84/root/etc"),
		stat("/dev/vda1", "/"),
		stat("/dev/vda1", "/usr/share/cagefs-skeleton/usr/lib64"),
		stat("/dev/vda1", "/usr"),
	}

	out := dedupeByDevice(in)

	if len(out) != 1 {
		t.Fatalf("want 1 deduped entry for /dev/vda1, got %d: %+v", len(out), out)
	}
	if out[0].labels.mountPoint != "/" {
		t.Fatalf("want representative mount %q, got %q", "/", out[0].labels.mountPoint)
	}
}

// TestDedupeByDevice_KubernetesCSIBindMountFanOut covers the other named
// trigger: Kubernetes CSI volume staging bind-mounts one block device into
// multiple paths (per node-agent hostfs prefix and per staging path). Real
// path shapes from a documented GKE case (same /dev/sde, no root mount
// among them, so the shortest staging path should win).
func TestDedupeByDevice_KubernetesCSIBindMountFanOut(t *testing.T) {
	const volID = "fd92a65917e16239faa6804bb34ba2adc94a7b432062ba3933ebae386eaa095e"
	in := []filesystemStats{
		stat("/dev/sde", "/hostfs/home/kubernetes/containerized_mounter/rootfs/var/lib/kubelet/plugins/kubernetes.io/csi/pd.csi.storage.gke.io/"+volID+"/globalmount"),
		stat("/dev/sde", "/hostfs/var/lib/kubelet/plugins/kubernetes.io/csi/pd.csi.storage.gke.io/"+volID+"/globalmount"),
		stat("/dev/sde", "/var/lib/kubelet/plugins/kubernetes.io/csi/pd.csi.storage.gke.io/"+volID+"/globalmount"),
	}

	out := dedupeByDevice(in)

	if len(out) != 1 {
		t.Fatalf("want 1 deduped entry for /dev/sde, got %d: %+v", len(out), out)
	}
	if out[0].labels.mountPoint != "/var/lib/kubelet/plugins/kubernetes.io/csi/pd.csi.storage.gke.io/"+volID+"/globalmount" {
		t.Fatalf("want shortest (non-hostfs-prefixed) mount, got %q", out[0].labels.mountPoint)
	}
}

// TestDedupeByDevice_NoRootPrefersShortestPath covers a device that never
// shows up mounted at "/" at all (e.g. a secondary data disk) - the
// shortest mountpoint should win as the closest-to-root representative.
func TestDedupeByDevice_NoRootPrefersShortestPath(t *testing.T) {
	in := []filesystemStats{
		stat("/dev/vdb1", "/mnt/data/deep/nested/path"),
		stat("/dev/vdb1", "/mnt/data"),
		stat("/dev/vdb1", "/mnt/data/deep"),
	}

	out := dedupeByDevice(in)

	if len(out) != 1 {
		t.Fatalf("want 1 deduped entry for /dev/vdb1, got %d: %+v", len(out), out)
	}
	if out[0].labels.mountPoint != "/mnt/data" {
		t.Fatalf("want shortest mount %q, got %q", "/mnt/data", out[0].labels.mountPoint)
	}
}

// TestDedupeByDevice_DistinctDevicesUntouched ensures real, independent
// filesystems (different devices) are never collapsed into each other.
func TestDedupeByDevice_DistinctDevicesUntouched(t *testing.T) {
	in := []filesystemStats{
		stat("/dev/vda1", "/"),
		stat("/dev/vda2", "/boot"),
		stat("/dev/vdb1", "/home"),
	}

	out := dedupeByDevice(in)

	if len(out) != 3 {
		t.Fatalf("want all 3 distinct-device entries kept, got %d: %+v", len(out), out)
	}
}

// TestDedupeByDevice_EmptyDevicePassesThrough ensures pseudo/virtual
// filesystems with no device string are never collapsed into each other,
// since an empty key would otherwise wrongly merge unrelated mounts.
func TestDedupeByDevice_EmptyDevicePassesThrough(t *testing.T) {
	in := []filesystemStats{
		stat("", "/proc"),
		stat("", "/sys"),
		stat("", "/dev/shm"),
	}

	out := dedupeByDevice(in)

	if len(out) != 3 {
		t.Fatalf("want all 3 empty-device entries passed through untouched, got %d: %+v", len(out), out)
	}
}
