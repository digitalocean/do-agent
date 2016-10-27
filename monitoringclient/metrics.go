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

package monitoringclient

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gogo/protobuf/proto"

	"github.com/digitalocean/do-agent/log"
	"github.com/digitalocean/do-agent/metrics"
)

const (
	authKeyHeader = "X-Auth-Key"
	// MetricsMasterURL is the address for metrics general server
	MetricsMasterURL = "https://master.sonar.digitalocean.com"

	// default push intervals in seconds for cases where the server does not specify a frequency
	defaultPushInterval   = 60
	jitterMin             = -15
	jitterMax             = 15
	jitterStdDev          = 3.2 // variance ~= 10
	pushIntervalHeaderKey = "X-Metric-Push-Interval"
	contentTypeHeader     = "Content-Type"
	httpTimeout           = 10 * time.Second
)

// DelimitedTelemetryContentType is the content type set on telemetry
// data responses in delimited protobuf format.
const DelimitedTelemetryContentType = `application/vnd.google.protobuf; proto=io.prometheus.client.MetricFamily; encoding=delimited`

// MonitoringMetricsClient interface describes available sonar metrics API calls
type MonitoringMetricsClient interface {
	SendMetrics() (int, error)

	Registry() metrics.Registry
}

// monitoringMetricsClientDroplet provide the interface for droplets to
//send metrics to Monitoring
type monitoringMetricsClientDroplet struct {
	url       string
	appKey    string
	dropletID int64
	r         metrics.Registry
}

//newMetricsClientDroplet creates a new monitoring metrics client with a specific region
func newMetricsClientDroplet(appKey string, dropletID int64, region, wharfURL string) monitoringMetricsClientDroplet {
	url := fmt.Sprintf("https://%s.sonar.digitalocean.com", region)
	if wharfURL != "" {
		url = wharfURL
		requireHTTPS = false
		log.Debugf("HTTPS requirement not enforced for overridden url: %s", url)
	}

	return monitoringMetricsClientDroplet{
		url:       url,
		appKey:    appKey,
		dropletID: dropletID,
		r:         metrics.NewRegistry(),
	}
}

// randomizedPushInterval returns an update interval w/ jitter applied.
func randomizedPushInterval() int {
	return defaultPushInterval +
		int(math.Max(math.Min(rand.NormFloat64()*jitterStdDev, jitterMax), jitterMin))
}

//SendMetrics sends metrics to monitoring server, the server returns how many seconds to wait until next push
func (s monitoringMetricsClientDroplet) SendMetrics() (int, error) {
	postURL := s.url + fmt.Sprintf("/v1/metrics/droplet_id/%d", s.dropletID)
	appKey := s.appKey
	nextPush := randomizedPushInterval()

	if s.r == nil {
		return nextPush, errors.New("no registry")
	}
	err := httpsCheck(postURL)
	if err != nil {
		return nextPush, err
	}

	// Collect all metrics.
	report := s.CreateReport()

	log.Debugf("Posting metrics to: %s", postURL)
	req, err := http.NewRequest("POST", postURL, bytes.NewBuffer(report))
	if err != nil {
		return nextPush, err
	}
	addUserAgentToHTTPRequest(req)
	req.Header.Set(contentTypeHeader, DelimitedTelemetryContentType)
	req.Header.Add(authKeyHeader, appKey)

	hc := http.Client{
		Timeout: httpTimeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout: httpTimeout,
			}).Dial,
			TLSHandshakeTimeout:   httpTimeout,
			ResponseHeaderTimeout: httpTimeout,
			DisableKeepAlives:     true,
		},
	}

	resp, err := hc.Do(req)
	if err != nil {
		return nextPush, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 202 {
		return nextPush,
			fmt.Errorf("unexpected status code %d while pushing to %s", resp.StatusCode, postURL)
	}

	receivedHeaders := resp.Header
	sendInterval, err := strconv.Atoi(receivedHeaders.Get(pushIntervalHeaderKey))
	if err != nil {
		sendInterval = nextPush
	}
	return sendInterval, nil
}

func (s monitoringMetricsClientDroplet) Registry() metrics.Registry {
	return s.r
}

//CreateMetricsClient creates a new metrics client
func CreateMetricsClient(appkey string, dropletID int64, region string, wharfURL string) (MonitoringMetricsClient, error) {
	return newMetricsClientDroplet(appkey, dropletID, region, wharfURL), nil
}

func (s monitoringMetricsClientDroplet) CreateReport() []byte {
	reporter := &prometheusReporter{
		metrics: make(map[string]*metrics.MetricFamily),
	}
	s.r.Report(reporter)

	var buf bytes.Buffer
	for _, m := range reporter.metrics {
		var err error
		if reporter.asText {
			err = proto.MarshalText(&buf, m)
		} else {
			_, err = appendDelimited(&buf, m)
		}
		if err != nil {
			log.Debugf("serialization error: %s", err)
		}
	}

	return buf.Bytes()
}

type prometheusReporter struct {
	metrics map[string]*metrics.MetricFamily
	asText  bool
}

func (p *prometheusReporter) Update(id metrics.MetricRef,
	value float64, labelValues ...string) {
	def, ok := id.(*metrics.Definition)
	if !ok {
		log.Debugf("unknown metric: %d", id)
		return
	}

	if len(labelValues) != len(def.MeasuredLabelKeys) {
		log.Debugf("label mismatch for metric: %s", def.Name)
		return
	}

	m := &metrics.Metric{}
	switch def.Type {
	case metrics.MetricType_COUNTER:
		m.Counter = &metrics.Counter{Value: value}
	case metrics.MetricType_GAUGE:
		m.Gauge = &metrics.Gauge{Value: value}
	}

	name := "sonar_" + def.Name
	fam, ok := p.metrics[name]
	if !ok {
		fam = &metrics.MetricFamily{
			Name: name,
			Type: def.Type,
		}
		p.metrics[name] = fam
	}
	fam.Metric = append(fam.Metric, m)

	if def.CommonLabels != nil {
		for k, v := range def.CommonLabels {
			m.Label = append(m.Label, &metrics.LabelPair{Name: k, Value: v})
		}
	}

	for i, v := range labelValues {
		m.Label = append(m.Label,
			&metrics.LabelPair{Name: def.MeasuredLabelKeys[i], Value: v})
	}
}

// appendDelimited appends a length-delimited protobuf message to the writer.
// Returns the number of bytes written, and any error.
func appendDelimited(out *bytes.Buffer, m proto.Message) (int, error) {
	buf, err := proto.Marshal(m)
	if err != nil {
		return 0, err
	}

	var delim [binary.MaxVarintLen32]byte
	len := binary.PutUvarint(delim[:], uint64(len(buf)))
	n, err := out.Write(delim[:len])
	if err != nil {
		return n, err
	}

	dn, err := out.Write(buf)
	return n + dn, err
}
