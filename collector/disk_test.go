// Copyright 2016 DigitalOcean
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package collector

import (
	"testing"

	"github.com/digitalocean/do-agent/procfs"
)

type stubDisker struct {
	NewDiskResultDisk []procfs.Disk
	NewDiskResultErr  error
}

func (s *stubDisker) NewDisk() ([]procfs.Disk, error) {
	return s.NewDiskResultDisk, s.NewDiskResultErr
}

// Verify that the stubDisker implements the procfs.Disker interface.
var _ procfs.Disker = (*stubDisker)(nil)

func TestRegisterDiskMetrics(t *testing.T) {
	testCases := []struct {
		label         string
		disker        *stubDisker
		expectedNames []string
	}{
		{"all_labels", &stubDisker{NewDiskResultErr: nil,
			NewDiskResultDisk: []procfs.Disk{
				procfs.Disk{
					MajorNumber:              uint64(1),
					MinorNumber:              uint64(2),
					DeviceName:               "fooDrive",
					ReadsCompleted:           uint64(3),
					ReadsMerged:              uint64(4),
					SectorsRead:              uint64(5),
					TimeSpentReading:         uint64(6),
					WritesCompleted:          uint64(7),
					WritesMerged:             uint64(8),
					SectorsWritten:           uint64(9),
					TimeSpendWriting:         uint64(10),
					IOInProgress:             uint64(11),
					TimeSpentDoingIO:         uint64(12),
					WeightedTimeSpentDoingIO: uint64(13),
				},
			},
		}, []string{
			"disk_io_now",
			"disk_io_time_ms",
			"disk_io_time_weighted",
			"disk_read_time_ms",
			"disk_reads_completed",
			"disk_reads_merged",
			"disk_sectors_read",
			"disk_sectors_written",
			"disk_write_time_ms",
			"disk_writes_completed",
			"disk_writes_merged",
			"disk_bytes_read",
			"disk_bytes_written",
		},
		},
		{"ignored_disk", &stubDisker{NewDiskResultErr: nil, NewDiskResultDisk: []procfs.Disk{procfs.Disk{DeviceName: "loop0"}}}, []string{
			"disk_io_now",
			"disk_io_time_ms",
			"disk_io_time_weighted",
			"disk_read_time_ms",
			"disk_reads_completed",
			"disk_reads_merged",
			"disk_sectors_read",
			"disk_sectors_written",
			"disk_write_time_ms",
			"disk_writes_completed",
			"disk_writes_merged",
			"disk_bytes_read",
			"disk_bytes_written",
		}},
	}

	for _, tc := range testCases {
		t.Run(tc.label, func(t *testing.T) {
			var actualNames []string

			r := &stubRegistry{}
			f := Filters{IncludeAll: true}
			RegisterDiskMetrics(r, tc.disker.NewDisk, f)

			for i := range r.RegisterNameOpts {
				actualNames = append(actualNames, r.RegisterNameOpts[i].Name)
			}

			testForMetricNames(t, tc.expectedNames, actualNames)

			if r.AddCollectorFunc == nil {
				t.Error("expected collector function, found none")
			}
		})
	}
}
