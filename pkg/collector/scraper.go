package collector

import (
	"bufio"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"github.com/prometheus/common/expfmt"

	"github.com/digitalocean/do-agent/internal/log"
	"github.com/digitalocean/do-agent/pkg/clients"
)

// NewScraper creates a new scraper to scrape metrics from the provided host
func NewScraper(name, metricsEndpoint string, extraMetricLabels []*dto.LabelPair, whitelist map[string]bool, timeout time.Duration) (*Scraper, error) {
	metricsEndpoint = strings.TrimRight(metricsEndpoint, "/")
	req, err := http.NewRequest("GET", metricsEndpoint, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create http request")
	}
	req.Header.Add("Accept", `text/plain;version=0.0.4;q=1,*/*;q=0.1`)
	req.Header.Add("Accept-Encoding", "gzip")
	req.Header.Set("User-Agent", "Prometheus/2.3.0")
	req.Header.Set("X-Prometheus-Scrape-Timeout-Seconds", fmt.Sprintf("%f", timeout.Seconds()))

	return &Scraper{
		req:               req,
		name:              name,
		extraMetricLabels: extraMetricLabels,
		whitelist:         whitelist,
		timeout:           timeout,
		client:            clients.NewHTTP(timeout),
		scrapeDurationDesc: prometheus.NewDesc(
			prometheus.BuildFQName(name, "scrape", "collector_duration_seconds"),
			fmt.Sprintf("%s: Duration of a collector scrape.", name),
			[]string{"collector"},
			nil,
		),
		scrapeSuccessDesc: prometheus.NewDesc(
			prometheus.BuildFQName(name, "scrape", "collector_success"),
			fmt.Sprintf("%s: Whether a collector succeeded.", name),
			[]string{"collector"},
			nil,
		),
	}, nil
}

// Scraper is a remote metric scraper that scrapes HTTP endpoints
type Scraper struct {
	timeout            time.Duration
	req                *http.Request
	client             *http.Client
	name               string
	whitelist          map[string]bool
	extraMetricLabels  []*dto.LabelPair
	scrapeDurationDesc *prometheus.Desc
	scrapeSuccessDesc  *prometheus.Desc
}

// readStream makes an HTTP request to the remote and returns the response body
// upon successful response
func (s *Scraper) readStream(ctx context.Context) (r io.ReadCloser, outerr error) {
	// close the reader if we return an error
	defer func() {
		if outerr == nil || r == nil {
			return
		}
		if err := r.Close(); err != nil {
			// This should not happen, but if it does it'll be nice
			// to know why we have a bunch of unclosed messages
			log.Error("failed to close stream on error: %+v", errors.WithStack(err))
		}
	}()

	resp, err := s.client.Do(s.req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "HTTP request failed")
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("server returned bad HTTP status %s", resp.Status)
	}

	if resp.Header.Get("Content-Encoding") != "gzip" {
		return resp.Body, nil
	}

	reader, err := gzip.NewReader(bufio.NewReader(resp.Body))
	return reader, errors.Wrap(err, "failed to create gzip reader")
}

// Describe describes this collector
func (s *Scraper) Describe(ch chan<- *prometheus.Desc) {
	ch <- s.scrapeDurationDesc
	ch <- s.scrapeSuccessDesc
}

// Collect collectrs metrics from the remote endpoint and reports them to ch
func (s *Scraper) Collect(ch chan<- prometheus.Metric) {
	var failed bool
	defer func(start time.Time) {
		dur := time.Since(start).Seconds()
		var success float64
		if !failed {
			success = 1
		}
		ch <- prometheus.MustNewConstMetric(s.scrapeDurationDesc, prometheus.GaugeValue, dur, s.Name())
		ch <- prometheus.MustNewConstMetric(s.scrapeSuccessDesc, prometheus.GaugeValue, success, s.Name())
	}(time.Now())

	ctx, cancel := context.WithTimeout(context.Background(), s.timeout)
	defer cancel()

	if err := s.scrape(ctx, ch); err != nil {
		failed = true
		log.Error("collection failed for %q: %v", s.Name(), err)
	}
}

func (s *Scraper) scrape(ctx context.Context, ch chan<- prometheus.Metric) (outerr error) {
	stream, err := s.readStream(ctx)
	if err != nil {
		return err
	}
	defer stream.Close()

	parsed, err := new(expfmt.TextParser).TextToMetricFamilies(stream)
	if err != nil {
		return errors.Wrapf(err, "parsing message failed")
	}

	for _, mf := range parsed {
		if s.FilterMetric(mf) {
			continue
		}
		convertMetricFamily(mf, ch, s.extraMetricLabels)
	}

	return nil
}

// Name returns the name of this scraper
func (s *Scraper) Name() string {
	return s.name
}

// FilterMetric returns true if the metric should be skipped (filtered out)
func (s *Scraper) FilterMetric(metricFamily *dto.MetricFamily) bool {
	if len(s.whitelist) == 0 { // if no whitelist treat all metrics as valid
		return false
	}

	return !s.whitelist[*metricFamily.Name]
}

// convertMetricFamily converts the dto metrics parsed from the expfmt package
// into the prometheus.Metrics required to pass over the channel
//
// this was copied and extended/refactored from github.com/prometheus/node_exporter
// see https://github.com/prometheus/node_exporter/blob/f56e8fcdf48ead56f1f149dbf1301ac028ef589b/collector/textfile.go#L63
// for more details
func convertMetricFamily(metricFamily *dto.MetricFamily, ch chan<- prometheus.Metric, extraLabels []*dto.LabelPair) {
	var valType prometheus.ValueType
	var val float64

	allLabelNames := getAllLabelNames(metricFamily, extraLabels)

	for _, metric := range metricFamily.Metric {
		names, values := getLabelNamesAndValues(metric, extraLabels, allLabelNames)

		metricType := metricFamily.GetType()
		switch metricType {
		case dto.MetricType_COUNTER:
			valType = prometheus.CounterValue
			val = metric.Counter.GetValue()

		case dto.MetricType_GAUGE:
			valType = prometheus.GaugeValue
			val = metric.Gauge.GetValue()

		case dto.MetricType_UNTYPED:
			valType = prometheus.UntypedValue
			val = metric.Untyped.GetValue()

		case dto.MetricType_SUMMARY:
			quantiles := map[float64]float64{}
			for _, q := range metric.Summary.Quantile {
				quantiles[q.GetQuantile()] = q.GetValue()
			}
			ch <- prometheus.MustNewConstSummary(
				prometheus.NewDesc(
					*metricFamily.Name,
					metricFamily.GetHelp(),
					names, nil,
				),
				metric.Summary.GetSampleCount(),
				metric.Summary.GetSampleSum(),
				quantiles, values...,
			)
		case dto.MetricType_HISTOGRAM:
			buckets := map[float64]uint64{}
			for _, b := range metric.Histogram.Bucket {
				buckets[b.GetUpperBound()] = b.GetCumulativeCount()
			}
			ch <- prometheus.MustNewConstHistogram(
				prometheus.NewDesc(
					*metricFamily.Name,
					metricFamily.GetHelp(),
					names, nil,
				),
				metric.Histogram.GetSampleCount(),
				metric.Histogram.GetSampleSum(),
				buckets, values...,
			)
		default:
			log.Error("unknown metric type %q", metricType.String())
			continue
		}
		if metricType == dto.MetricType_GAUGE || metricType == dto.MetricType_COUNTER || metricType == dto.MetricType_UNTYPED {
			ch <- prometheus.MustNewConstMetric(
				prometheus.NewDesc(
					*metricFamily.Name,
					metricFamily.GetHelp(),
					names, nil,
				),
				valType, val, values...,
			)
		}
	}
}

// getLabelNamesAndValues returns a slice of label names and a slice of label values from the metric and extra labels.
func getLabelNamesAndValues(metric *dto.Metric, extraLabels []*dto.LabelPair, allLabelNames map[string]struct{}) ([]string, []string) {
	labels := metric.GetLabel()
	if extraLabels != nil {
		labels = append(labels, extraLabels...)
	}
	var names []string
	var values []string
	for _, label := range labels {
		names = append(names, label.GetName())
		values = append(values, label.GetValue())
	}
	for k := range allLabelNames {
		present := false
		for _, name := range names {
			if k == name {
				present = true
				break
			}
		}
		if !present {
			names = append(names, k)
			values = append(values, "")
		}
	}
	return names, values
}

// getAllLabelNames returns the map of all label names from the metric family including any extra labels provided.
func getAllLabelNames(metricFamily *dto.MetricFamily, extraLabels []*dto.LabelPair) map[string]struct{} {
	allLabelNames := map[string]struct{}{}
	for _, metric := range metricFamily.Metric {
		labels := metric.GetLabel()
		if extraLabels != nil {
			labels = append(labels, extraLabels...)
		}
		for _, label := range labels {
			if _, ok := allLabelNames[label.GetName()]; !ok {
				allLabelNames[label.GetName()] = struct{}{}
			}
		}
	}
	return allLabelNames
}
