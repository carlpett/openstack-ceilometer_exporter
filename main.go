package main

import (
	"flag"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/DSpeichert/gophercloud/openstack"
	"github.com/DSpeichert/gophercloud/openstack/telemetry/v2/meters"
	"github.com/rackspace/gophercloud"
	"github.com/rackspace/gophercloud/openstack/compute/v2/servers"
	"github.com/rackspace/gophercloud/openstack/networking/v2/extensions/lbaas/pools"
	"github.com/rackspace/gophercloud/pagination"

	"github.com/prometheus/client_golang/prometheus"

	log "github.com/Sirupsen/logrus"
	"github.com/ryanuber/go-glob"
)

/*
  TODOs:
  - Auth without env?
  - Split metric types (HW/Resources/...) (?)
  - Support for meter/foo/statistics for some types?
  - Multiple scrapers
  - Split to multiple files
  - Calculated metrics (eg count of rules in firewall policy)
  - Timeout?
*/

func init() {
	flag.Parse()

	parsedLevel, err := log.ParseLevel(*rawLevel)
	if err != nil {
		log.Fatal(err)
	}
	logLevel = parsedLevel

	enabledMetrics = strings.Split(*rawEnabledMetrics, ",")
	disabledMetrics = strings.Split(*rawDisabledMetrics, ",")
}

const (
	namespace             = "openstack_ceilometer"
	defaultEnabledMetrics = "*"
)

var (
	logLevel           = log.InfoLevel
	rawLevel           = flag.String("log-level", "info", "log level")
	bindAddr           = flag.String("bind-addr", ":9181", "bind address for the metrics server")
	metricsPath        = flag.String("metrics-path", "/metrics", "path to metrics endpoint")
	maxResults         = flag.Int("max-results", 100, "maximum number of results to fetch for any metric")
	maxMetricAge       = flag.Duration("max-metric-age", 5*time.Minute, "maximum age of metrics to retrieve")
	enabledMetrics     []string
	rawEnabledMetrics  = flag.String("enabled-metrics", defaultEnabledMetrics, "comma-separated list of metrics to enable (supports globbing)")
	disabledMetrics    []string
	rawDisabledMetrics = flag.String("disabled-metrics", "", "comma-separated list of metrics to disable (supports globbing)")
	listMetrics        = flag.Bool("list-metrics", false, "show list of metrics and exit")
)

func shouldUseMetric(metric string) bool {
	for _, enabled := range enabledMetrics {
		if glob.Glob(enabled, metric) {
			for _, disabled := range disabledMetrics {
				if glob.Glob(disabled, metric) {
					return false
				}
			}
			return true
		}
	}
	return false
}

func main() {
	log.SetLevel(logLevel)
	if *listMetrics {
		displayMetricsList()
		return
	}

	prometheus.MustRegister(NewCeilometerCollector())

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<html>
             <head><title>Openstack Ceilometer Exporter</title></head>
             <body>
             <h1>Openstack Ceilometer Exporter</h1>
             <p><a href='` + *metricsPath + `'>Metrics</a></p>
             </body>
             </html>`))
	})
	log.Infof("Starting metric server on %v%v", *bindAddr, *metricsPath)
	log.Fatal(http.ListenAndServe(*bindAddr, nil))
}

func displayMetricsList() {
	metrics := *getMetrics(nil)
	availableMetrics := make([]string, 0, len(metrics))
	for name, _ := range metrics {
		availableMetrics = append(availableMetrics, name)
	}
	sort.Strings(availableMetrics)

	for _, metric := range availableMetrics {
		fmt.Println(metric)
	}
}

type LookupService struct {
	poolNameCache map[string]string
	networkClient *gophercloud.ServiceClient

	instanceNameCache map[string]string
	serverClient      *gophercloud.ServiceClient
}

func NewLookupService(provider *gophercloud.ProviderClient) LookupService {
	networkClient, err := openstack.NewNetworkV2(provider, gophercloud.EndpointOpts{})
	if err != nil {
		panic(err)
	}

	serverClient, err := openstack.NewComputeV2(provider, gophercloud.EndpointOpts{})
	if err != nil {
		panic(err)
	}

	log.Debug("Populating guid lookup caches")

	poolNameCache := make(map[string]string)
	poolPager := pools.List(networkClient, pools.ListOpts{})
	poolPager.EachPage(func(page pagination.Page) (bool, error) {
		poolList, err := pools.ExtractPools(page)
		if err != nil {
			return false, err
		}
		for _, pool := range poolList {
			poolNameCache[pool.ID] = pool.Name
		}
		return true, nil
	})

	serverNameCache := make(map[string]string)
	serverPager := servers.List(serverClient, servers.ListOpts{})
	serverPager.EachPage(func(page pagination.Page) (bool, error) {
		serverList, err := servers.ExtractServers(page)
		if err != nil {
			return false, err
		}
		for _, server := range serverList {
			serverNameCache[server.ID] = server.Name
		}
		return true, nil
	})

	log.Debugf("Finished populating caches. %d pools and %d instances prepared.", len(poolNameCache), len(serverNameCache))

	return LookupService{
		networkClient:     networkClient,
		poolNameCache:     poolNameCache,
		serverClient:      serverClient,
		instanceNameCache: serverNameCache,
	}
}

func (this *LookupService) lookupPool(poolId string) string {
	if poolId == "" {
		return "UNKNOWN"
	}

	var name string
	if name, ok := this.poolNameCache[poolId]; ok {
		return name
	}

	result := pools.Get(this.networkClient, poolId)
	pool, err := result.Extract()
	if err != nil {
		log.Warnf("Failure while looking up pool id %q", poolId)
		name = "UNKNOWN"
	} else {
		name = pool.Name
	}
	this.poolNameCache[poolId] = name
	return name
}

func (this *LookupService) lookupInstance(instanceId string) string {
	if instanceId == "" {
		return "UNKNOWN"
	}

	var name string
	if name, ok := this.instanceNameCache[instanceId]; ok {
		return name
	}

	result := servers.Get(this.serverClient, instanceId)
	instance, err := result.Extract()
	if err != nil {
		log.Warnf("Failure while looking up instance id %q", instanceId)
		name = "UNKNOWN"
	} else {
		name = instance.Name
	}
	this.instanceNameCache[instanceId] = name
	return name
}

func makeFQName(metric string) string {
	return fmt.Sprintf("%s_%s", namespace, metric)
}

func NewCeilometerCollector() *ceilometerCollector {
	opts, err := openstack.AuthOptionsFromEnv()
	if err != nil {
		panic(err)
	}
	provider, err := openstack.AuthenticatedClient(opts)
	if err != nil {
		panic(err)
	}

	client, err := openstack.NewTelemetryV2(provider, gophercloud.EndpointOpts{})
	if err != nil {
		panic(err)
	}

	lookupSvc := NewLookupService(provider)

	allMetrics := *getMetrics(&lookupSvc)
	filteredMetrics := make(map[string]ceilometerMetric)
	for name, metric := range allMetrics {
		if shouldUseMetric(name) {
			filteredMetrics[name] = metric
		}
	}

	return &ceilometerCollector{
		metrics: filteredMetrics,
		metaMetrics: map[string]*prometheus.Desc{
			"scrapeSuccess":    prometheus.NewDesc(makeFQName("metric_scrape_success"), "Indicates if the metric was successfully scraped", []string{"metric"}, nil),
			"scrapeDuration":   prometheus.NewDesc(makeFQName("metric_scrape_duration_ns"), "The time taken to scrape the metric", []string{"metric"}, nil),
			"scrapeResultSize": prometheus.NewDesc(makeFQName("metric_scrape_result_size"), "Number of results returned by the metric query", []string{"metric"}, nil),

			"totalScrapeDuration": prometheus.NewDesc(makeFQName("total_scrape_duration_ns"), "Time taken for entire scrape", nil, nil),
		},
		client: client,
	}
}

type ceilometerCollector struct {
	client      *gophercloud.ServiceClient
	metrics     map[string]ceilometerMetric
	metaMetrics map[string]*prometheus.Desc
}
type ceilometerMetric struct {
	desc          *prometheus.Desc
	extractLabels func(*meters.OldSample) []string
}

func (c *ceilometerCollector) Describe(ch chan<- *prometheus.Desc) {
	log.Debugf("Sending %d metrics descriptions", len(c.metrics)+len(c.metaMetrics))
	for _, metric := range c.metrics {
		ch <- metric.desc
	}
	for _, metric := range c.metaMetrics {
		ch <- metric
	}
}

func (c *ceilometerCollector) Collect(ch chan<- prometheus.Metric) {
	t := time.Now()
	result := make(chan scrapeStats)
	defer close(result)
	for resourceLabel, metric := range c.metrics {
		go scrape(resourceLabel, metric, c.client, ch, result)
	}
	for _ = range c.metrics {
		scrapeStats := <-result
		ch <- prometheus.MustNewConstMetric(c.metaMetrics["scrapeSuccess"], prometheus.GaugeValue, btof(scrapeStats.success), scrapeStats.resourceLabel)
		ch <- prometheus.MustNewConstMetric(c.metaMetrics["scrapeDuration"], prometheus.GaugeValue, float64(scrapeStats.duration.Nanoseconds()), scrapeStats.resourceLabel)
		ch <- prometheus.MustNewConstMetric(c.metaMetrics["scrapeResultSize"], prometheus.GaugeValue, float64(scrapeStats.resultSize), scrapeStats.resourceLabel)
	}

	ch <- prometheus.MustNewConstMetric(c.metaMetrics["totalScrapeDuration"], prometheus.GaugeValue, float64(time.Since(t).Nanoseconds()))
}

func btof(b bool) float64 {
	if b {
		return 1.0
	} else {
		return 0.0
	}
}

type scrapeStats struct {
	resourceLabel string
	success       bool
	duration      time.Duration
	resultSize    int
}

func sendStats(ch chan<- scrapeStats, stats *scrapeStats) {
	ch <- *stats
}
func registerDuration(start time.Time, stats *scrapeStats) {
	stats.duration = time.Since(start)
}

func scrape(resourceLabel string, metric ceilometerMetric, client *gophercloud.ServiceClient, ch chan<- prometheus.Metric, result chan<- scrapeStats) {
	t := time.Now()
	stats := scrapeStats{resourceLabel: resourceLabel}
	defer sendStats(result, &stats)
	defer registerDuration(t, &stats)

	query := meters.ShowOpts{
		QueryField: "timestamp",
		QueryOp:    "gt",
		QueryValue: time.Now().UTC().Add(-*maxMetricAge).Format("2006-01-02T15:04:05"),
		Limit:      *maxResults,
	}
	log.Debugf("Querying for %v: %v", resourceLabel, query)
	results := meters.Show(client, resourceLabel, query)
	data, err := results.Extract()
	if err != nil {
		log.Warnf("Failed to scrape Ceilometer resource %q", resourceLabel)
		return
	}
	if len(data) == 0 {
		log.Warnf("Query for %v returned no results!", resourceLabel)
		stats.success = true // The query itself was successful, even though no results were produced
		return
	}
	if len(data) == *maxResults {
		log.Warnf("Query for %v returned max number of results (%d), data may be truncated", resourceLabel, *maxResults)
	}
	initialLen := len(data)
	data = deduplicate(data)
	log.Debugf("Query for %s returned %d results, %d remain after deduplication", resourceLabel, initialLen, len(data))
	stats.resultSize = len(data)

	for _, sample := range data {
		ch <- sampleToMetric(&sample, metric)
	}

	stats.success = true
}

func deduplicate(samples []meters.OldSample) []meters.OldSample {
	unique := make([]meters.OldSample, 0, len(samples))
	seen := make(map[string]bool)
	for _, sample := range samples {
		if _, ok := seen[sample.ResourceId]; !ok {
			seen[sample.ResourceId] = true
			unique = append(unique, sample)
		}
	}
	return unique
}

func sampleToMetric(sample *meters.OldSample, metric ceilometerMetric) prometheus.Metric {
	var valueType prometheus.ValueType
	switch sample.Type {
	case "gauge":
		valueType = prometheus.GaugeValue
	case "cumulative":
		valueType = prometheus.CounterValue

	default:
		log.Debugf("Unknown sample type %v in query for %v", sample.Type, sample.Name)
		valueType = prometheus.UntypedValue
	}

	value := float64(sample.Volume)

	return prometheus.MustNewConstMetric(metric.desc, valueType, value, metric.extractLabels(sample)...)
}
