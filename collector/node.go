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
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/digitalocean/do-agent/config"
	"github.com/digitalocean/do-agent/log"
	"github.com/digitalocean/do-agent/metrics"
	"github.com/digitalocean/do-agent/procfs"
)

const nodeSystem = "node"

type osReleaseFunc func() (procfs.OSRelease, error)

func kernelVersion(f osReleaseFunc) string {
	version, err := f()
	if err != nil {
		log.Debugf("Unable to collect kernel version: %s", err)
		return "Unavailable kernel version"
	}
	return string(version)
}

//RegisterNodeMetrics creates a reference to a NodeCollector.
func RegisterNodeMetrics(r metrics.Registry, fn osReleaseFunc) {
	labels := map[string]string{
		"os":                  runtime.GOOS,
		"architecture":        runtime.GOARCH,
		"sonar_agent_version": config.Version(),
		"build":               config.Build(),
		"kernel":              kernelVersion(fn),
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		log.Debug("Unable to collect interface IP Addresses")
		return
	}

	ipCount := 0
	for _, addr := range addrs {
		var ip net.IP
		switch v := addr.(type) {
		case *net.IPNet:
			ip = v.IP
		case *net.IPAddr:
			ip = v.IP
		default:
			continue
		}
		if ip.IsLoopback() {
			continue
		}
		ipCount++
		labels[fmt.Sprintf("ipaddress%d", ipCount)] = ip.String()
	}

	info := r.Register(nodeSystem+"_info",
		metrics.WithCommonLabels(labels),
		metrics.WithMeasuredLabels("host_name"))

	r.AddCollector(func(r metrics.Reporter) {
		hostName, err := os.Hostname()
		if err != nil {
			log.Debugf("Unable to collect hostname: %s", err)
			hostName = "Unavailable Hostname"
		}

		r.Update(info, 0, hostName)
	})
}
