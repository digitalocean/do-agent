/*
Package tsclient provides a common client for sending metrics to the DO timeseries system.

The timeseries system is a push-based system where metrics are submitted in batches
via the SendMetrics method at fixed time intervals. Metrics are submitted to the wharf
server.

Wharf responds with a rate-limit value which the client must wait that many seconds
or longer before submitting the next batch of metrics -- this is exposed via the WaitDuration()
method.

*/
package tsclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/golang/snappy"
	"github.com/pkg/errors"

	"github.com/digitalocean/do-agent/internal/log"
	"github.com/digitalocean/do-agent/pkg/clients/tsclient/structuredstream"
)

const (
	binaryContentType = "application/timeseries-binary-0"
	jsonContentType   = "application/json"
	userAgentHeader   = "User-Agent"
	authKeyHeader     = "X-Auth-Key"
	contentTypeHeader = "Content-Type"

	defaultWaitIntervalSeconds = 60
	defaultMaxBatchSize        = 1000
	defaultMaxMetricLength     = 512
	maxWaitInterval            = 10 * time.Minute
)

// Client is an interface for sending batches of metrics
type Client interface {
	AddMetric(def *Definition, value float64, labels ...string) error
	AddMetricWithTime(def *Definition, t time.Time, value float64, labels ...string) error
	Flush() error
	WaitDuration() time.Duration
	MaxBatchSize() int
	MaxMetricLength() int
	ResetWaitTimer()
}

// HTTPClient is used to send metrics via http
type HTTPClient struct {
	httpClient               *http.Client
	userAgent                string
	metadataEndpoint         string
	radarEndpoint            string
	wharfEndpoints           []string
	wharfEndpointSSLHostname string
	lastFlushAttempt         time.Time
	lastFlushConnection      time.Time
	waitIntervalSeconds      int32
	maxBatchSize             int32
	maxMetricLength          int32
	numConsecutiveFailures   int
	bootstrapRequired        bool
	trusted                  bool
	lastSend                 map[string]int64
	isZeroTime               bool

	// variables only used when trusted
	appName string
	appKey  string

	// variables only used when non-trusted
	dropletID string
	region    string

	buf *bytes.Buffer
	w   *snappy.Writer
}

// ClientOptions are client options
type ClientOptions struct {
	UserAgent                string
	WharfEndpoints           []string
	WharfEndpointSSLHostname string
	AppName                  string
	AppKey                   string
	MetadataEndpoint         string
	RadarEndpoint            string
	Timeout                  time.Duration
	IsTrusted                bool
	MaxBatchSize             int
	MaxMetricLength          int
}

// ClientOptFn allows for overriding options
type ClientOptFn func(*ClientOptions)

// WithWharfEndpoint overrides the default wharf endpoint, this option must be set when WithTrustedAppKey is used.
func WithWharfEndpoint(endpoint string) ClientOptFn {
	return WithWharfEndpoints([]string{endpoint})
}

// WithWharfEndpoints overrides the default wharf endpoint, this option must be set when WithTrustedAppKey is used.
func WithWharfEndpoints(endpoints []string) ClientOptFn {
	return func(o *ClientOptions) {
		o.WharfEndpoints = endpoints
	}
}

// WithWharfEndpointSSLHostname overrides the default wharf endpoint, this option must be set when WithTrustedAppKey is used.
func WithWharfEndpointSSLHostname(hostname string) ClientOptFn {
	return func(o *ClientOptions) {
		o.WharfEndpointSSLHostname = hostname
	}
}

// WithMetadataEndpoint overrides the default metadata endpoint, this option is only applicable to non-trusted clients (i.e. running on a customer droplet).
func WithMetadataEndpoint(endpoint string) ClientOptFn {
	return func(o *ClientOptions) {
		o.MetadataEndpoint = endpoint
	}
}

// WithRadarEndpoint overrides the default radar endpoint, this option is only applicable to non-trusted clients (i.e. running on a customer droplet).
func WithRadarEndpoint(endpoint string) ClientOptFn {
	return func(o *ClientOptions) {
		o.RadarEndpoint = endpoint
	}
}

// WithTimeout overrides the default wharf endpoint
func WithTimeout(timeout time.Duration) ClientOptFn {
	return func(o *ClientOptions) {
		o.Timeout = timeout
	}
}

// WithUserAgent overrides the http user agent
func WithUserAgent(s string) ClientOptFn {
	return func(o *ClientOptions) {
		o.UserAgent = s
	}
}

// WithTrustedAppKey disables metadata authentication; trusted apps can override the host_id and user_id tags.
func WithTrustedAppKey(appName, appKey string) ClientOptFn {
	return func(o *ClientOptions) {
		o.AppName = appName
		o.AppKey = appKey
		o.IsTrusted = true
	}
}

// WithDefaultLimits set default metric limits. These will always be overridden by the server after first write
func WithDefaultLimits(maxBatchSize, maxMetricLength int) ClientOptFn {
	return func(o *ClientOptions) {
		o.MaxBatchSize = maxBatchSize
		o.MaxMetricLength = maxMetricLength
	}
}

// New creates a new client
func New(opts ...ClientOptFn) Client {
	opt := &ClientOptions{
		UserAgent:        "tsclient-unknown",
		Timeout:          10 * time.Second,
		MetadataEndpoint: "http://169.254.169.254/metadata",
		RadarEndpoint:    "https://169.254.169.254",
	}

	for _, fn := range opts {
		fn(opt)
	}

	if opt.MaxMetricLength == 0 {
		opt.MaxMetricLength = defaultMaxMetricLength
	}

	if opt.MaxBatchSize == 0 {
		opt.MaxBatchSize = defaultMaxBatchSize
	}

	var tlsConfig tls.Config
	if opt.WharfEndpointSSLHostname != "" {
		tlsConfig.ServerName = opt.WharfEndpointSSLHostname
	}

	httpClient := &http.Client{
		Timeout: opt.Timeout,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout: opt.Timeout,
			}).Dial,
			TLSHandshakeTimeout:   opt.Timeout,
			ResponseHeaderTimeout: opt.Timeout,
			DisableKeepAlives:     true,
			TLSClientConfig:       &tlsConfig,
		},
	}

	if opt.IsTrusted {
		if len(opt.WharfEndpoints) == 0 {
			panic("WithWharfEndpoint() must be used WithTrustedAppKey")
		}
		if opt.AppName == "" {
			panic("appname must be set")
		}
	}

	return &HTTPClient{
		userAgent:                opt.UserAgent,
		metadataEndpoint:         opt.MetadataEndpoint,
		radarEndpoint:            opt.RadarEndpoint,
		wharfEndpoints:           opt.WharfEndpoints,
		wharfEndpointSSLHostname: opt.WharfEndpointSSLHostname,
		appName:                  opt.AppName,
		appKey:                   opt.AppKey,
		httpClient:               httpClient,
		waitIntervalSeconds:      defaultWaitIntervalSeconds,
		maxBatchSize:             int32(opt.MaxBatchSize),
		maxMetricLength:          int32(opt.MaxMetricLength),
		bootstrapRequired:        true,
		trusted:                  opt.IsTrusted,
		lastSend:                 map[string]int64{},
	}
}

func (c *HTTPClient) bootstrapFromMetadata() error {
	var err error

	if c.trusted {
		return nil
	}

	c.dropletID, err = c.GetDropletID()
	if err != nil {
		return err
	}
	log.Debug("droplet ID: %s", c.dropletID)

	c.region, err = c.GetRegion()
	if err != nil {
		return err
	}
	log.Debug("region: %s", c.region)

	authToken, err := c.GetAuthToken()
	if err != nil {
		return err
	}
	log.Debug("auth token: %s", truncate(authToken, 5))

	appKey, err := c.GetAppKey(authToken)
	if err != nil {
		return err
	}
	c.appKey = appKey
	log.Debug("appkey: %s", truncate(c.appKey, 5))

	return nil
}

// url returns a potentially randomized endpoint to send data to
// the url must constantly be randomized; otherwise the cache across all wharf endpoints
// will be skewed (i.e. only a single node will know about the droplet -> user ID lookups)
// and when a restart/failure finally happens, then a different wharf endpoint will be picked,
// and it wont have anything in its cache.
func (c *HTTPClient) url() string {
	if c.trusted {
		if len(c.wharfEndpoints) == 0 {
			panic("trusted app with no wharf endpoints; shouldnt happen")
		}
		if c.appName == "" {
			panic("appname not defined; shouldnt happen")
		}
		endpoint := c.wharfEndpoints[rand.Intn(len(c.wharfEndpoints))]
		return fmt.Sprintf("%s/v1/metrics/trusted/%s", endpoint, c.appName)
	}

	endpoint := "http://169.254.169.254"
	if len(c.wharfEndpoints) > 0 {
		endpoint = c.wharfEndpoints[rand.Intn(len(c.wharfEndpoints))]
	}
	return fmt.Sprintf("%s/v1/metrics/droplet_id/%s", endpoint, c.dropletID)
}

// WaitDuration returns the duration before the next batch of metrics will be accepted
func (c *HTTPClient) WaitDuration() time.Duration {
	d := time.Since(c.lastFlushAttempt)
	wi := time.Second * time.Duration(atomic.LoadInt32(&c.waitIntervalSeconds))
	if d < wi {
		return wi - d
	}
	return 0
}

// MaxBatchSize returns the maximum amount of metrics that may be sent per batch
func (c *HTTPClient) MaxBatchSize() int {
	return int(atomic.LoadInt32(&c.maxBatchSize))
}

// MaxMetricLength is the maximum length of a metric that may be sent (all labels and values combined)
func (c *HTTPClient) MaxMetricLength() int {
	return int(atomic.LoadInt32(&c.maxMetricLength))
}

// AddMetric adds a metric to the batch
func (c *HTTPClient) AddMetric(def *Definition, value float64, labels ...string) error {
	return c.addMetricWithMSEpochTime(def, 0, value, labels...)
}

// AddMetricWithTime adds a metric to the batch
func (c *HTTPClient) AddMetricWithTime(def *Definition, t time.Time, value float64, labels ...string) error {
	ms := t.UTC().UnixNano() / int64(time.Millisecond)
	return c.addMetricWithMSEpochTime(def, ms, value, labels...)
}

func (c *HTTPClient) addMetricWithMSEpochTime(def *Definition, ms int64, value float64, labels ...string) error {
	isZeroTime := bool(ms == 0)
	if c.buf == nil {
		c.buf = new(bytes.Buffer)
		c.w = snappy.NewBufferedWriter(c.buf)
		c.lastSend = map[string]int64{}
		c.isZeroTime = isZeroTime
	} else {
		if isZeroTime != c.isZeroTime {
			panic("client support for AddMetrics and AddMetricWithTime is mutually exclusive")
		}
	}
	lfm, err := GetLFM(def, labels)
	if err != nil {
		return errors.Wrap(err, "failed to get LFM")
	}

	if !isZeroTime {
		// ensure sufficient time between reported metric values
		if lastSend, ok := c.lastSend[lfm]; ok && (time.Duration(ms-lastSend)*time.Millisecond) < c.WaitDuration() {
			return errors.WithStack(ErrSendTooFrequent)
		}
		c.lastSend[lfm] = ms
	}

	writer := structuredstream.NewWriter(c.w)
	writer.WriteUint16PrefixedString(lfm)
	writer.Write(ms)
	writer.Write(value)
	if err := writer.Error(); err != nil {
		log.Error("failed to write: %+v", err)
		return errors.WithStack(ErrWriteFailure)
	}
	return nil
}

func (c *HTTPClient) clearBufferedMetrics() {
	c.buf = nil

	// clean lastSend (potential memory leak otherwise)
	nowMS := time.Now().UTC().UnixNano() / int64(time.Millisecond)
	for lfm, t := range c.lastSend {
		if (nowMS - t) > 60*60*1000 {
			delete(c.lastSend, lfm)
		}
	}
}

// ResetWaitTimer causes the wait duration timer to reset
func (c *HTTPClient) ResetWaitTimer() {
	c.lastFlushAttempt = time.Now()
}

// Flush sends the batch of metrics to wharf
func (c *HTTPClient) Flush() error {
	now := time.Now()
	if now.Sub(c.lastFlushAttempt) < c.WaitDuration() {
		return ErrFlushTooFrequent
	}
	c.lastFlushAttempt = now

	if c.numConsecutiveFailures > 3 {
		timeSinceLastConnection := now.Sub(c.lastFlushConnection)
		requiredWait := time.Minute * time.Duration(c.numConsecutiveFailures+rand.Intn(3))
		if requiredWait > maxWaitInterval {
			requiredWait = maxWaitInterval
		}
		if timeSinceLastConnection < requiredWait {
			return ErrCircuitBreaker
		}
	}

	if c.buf == nil {
		return nil
	}

	c.lastFlushConnection = now

	if c.bootstrapRequired || c.numConsecutiveFailures > 60 {
		if err := c.bootstrapFromMetadata(); err != nil {
			c.numConsecutiveFailures++
			return err
		}
		c.bootstrapRequired = false
	}

	err := c.w.Flush()
	if err != nil {
		return err
	}

	url := c.url()
	log.Debug("sending metrics to %s", url)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(c.buf.Bytes()))
	if err != nil {
		c.numConsecutiveFailures++
		if c.isZeroTime {
			c.clearBufferedMetrics()
		}
		return err
	}

	req.Header.Add(userAgentHeader, c.userAgent)
	if c.wharfEndpointSSLHostname != "" {
		req.Host = c.wharfEndpointSSLHostname
	}
	req.Header.Set(contentTypeHeader, binaryContentType)
	req.Header.Add(authKeyHeader, c.appKey)

	resp, err := c.httpClient.Do(req.WithContext(context.Background()))
	if err != nil {
		c.numConsecutiveFailures++
		if c.isZeroTime {
			c.clearBufferedMetrics()
		}
		return err
	}
	contentType := resp.Header.Get(contentTypeHeader)
	if contentType == jsonContentType {
		defer c.handleSonarResponse(resp.Body)
	}
	if resp.StatusCode != http.StatusAccepted {
		c.numConsecutiveFailures++
		if c.isZeroTime {
			c.clearBufferedMetrics()
		}
		return &UnexpectedHTTPStatusError{StatusCode: resp.StatusCode}
	}

	c.numConsecutiveFailures = 0
	c.clearBufferedMetrics()
	return nil
}

// handleSonarResponse reads sonar response messages and parses limits, setting them for future usages
func (c *HTTPClient) handleSonarResponse(r io.ReadCloser) {
	defer r.Close()
	res, err := readBody(r)
	if err != nil {
		log.Error("failed to read response body of sonar message: +%v", err)
	} else {
		if res.FrequencySeconds != 0 {
			atomic.StoreInt32(&c.waitIntervalSeconds, res.FrequencySeconds)
		}
		if res.MaxMetricLength != 0 {
			atomic.StoreInt32(&c.maxMetricLength, res.MaxMetricLength)
		}
		if res.MaxBatchSize != 0 {
			atomic.StoreInt32(&c.maxBatchSize, res.MaxBatchSize)
		}
	}
}

// GetWaitInterval returns the wait interval between metrics
func (c *HTTPClient) GetWaitInterval() time.Duration {
	return time.Second * time.Duration(atomic.LoadInt32(&c.waitIntervalSeconds))
}

// GetDropletID returns the droplet ID
func (c *HTTPClient) GetDropletID() (string, error) {
	return c.httpGet(fmt.Sprintf("%s/v1/id", c.metadataEndpoint), "")
}

// GetRegion returns the region
func (c *HTTPClient) GetRegion() (string, error) {
	return c.httpGet(fmt.Sprintf("%s/v1/region", c.metadataEndpoint), "")
}

// GetAuthToken returns an auth token
func (c *HTTPClient) GetAuthToken() (string, error) {
	return c.httpGet(fmt.Sprintf("%s/v1/auth-token", c.metadataEndpoint), "")
}

// GetAppKey returns the appkey
func (c *HTTPClient) GetAppKey(authToken string) (string, error) {
	body, err := c.httpGet(fmt.Sprintf("%s/v1/appkey/droplet-auth-token", c.radarEndpoint), authToken)
	if err != nil {
		return "", err
	}

	var appKey string
	err = json.Unmarshal([]byte(body), &appKey)
	if err != nil {
		return "", err
	}

	return appKey, nil
}

func truncate(str string, num int) string {
	s := str
	if len(str) > num {
		if num > 3 {
			num -= 3
		}
		s = str[0:num] + "*******"
	}
	return s
}

func (c *HTTPClient) httpGet(url, authToken string) (string, error) {
	log.Debug("HTTP GET %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	if authToken != "" {
		authValue := "DOMETADATA " + authToken
		req.Header.Add("Authorization", authValue)
		log.Debug("Authorization: %s", truncate(authValue, 15))
	}

	resp, err := c.httpClient.Do(req.WithContext(context.Background()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Debug("got status code %d while fetching %s (auth token: %s)", resp.StatusCode, url, truncate(authToken, 5))
		return "", &UnexpectedHTTPStatusError{StatusCode: resp.StatusCode}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

//{"success":true,"frequency":60,"max_metrics":1000,"max_lfm":512}
type sonarResponse struct {
	Success          bool  `json:"success"`
	FrequencySeconds int32 `json:"frequency"`
	MaxBatchSize     int32 `json:"max_metrics"`
	MaxMetricLength  int32 `json:"max_lfm"`
}

func readBody(r io.Reader) (sonarResponse, error) {
	var res sonarResponse
	return res, json.NewDecoder(r).Decode(&res)
}
