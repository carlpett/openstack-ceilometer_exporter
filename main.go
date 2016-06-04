package main

import (
	"flag"
	"net/http"
	"strings"
	"time"

	"github.com/DSpeichert/gophercloud/openstack"
	"github.com/DSpeichert/gophercloud/openstack/telemetry/v2/meters"
	upstream "github.com/rackspace/gophercloud"

	"github.com/prometheus/client_golang/prometheus"

	log "github.com/Sirupsen/logrus"
)

/*
  TODOs:
  - Flags
  - Split metric types (HW/Resources/...) (?)
  - Support for meter/foo/statistics for some types?
  - Scrape-stats (scrape time, success, etc) (split scrape time per metric?)
  - Multiple scrapers
  - Split to multiple files
  - Calculated metrics (eg count of rules in firewall policy)
  - Timeout?
*/

/*
Types:
  cpu
  cpu_util
  disk.allocation
  disk.capacity
  disk.device.allocation
  disk.device.capacity
  disk.device.read.bytes
  disk.device.read.bytes.rate
  disk.device.read.requests
  disk.device.read.requests.rate
  disk.device.usage
  disk.device.write.bytes
  disk.device.write.bytes.rate
  disk.device.write.requests
  disk.device.write.requests.rate
  disk.ephemeral.size
  disk.read.bytes
  disk.read.bytes.rate
  disk.read.requests
  disk.read.requests.rate
  disk.root.size
  disk.usage
  disk.write.bytes
  disk.write.bytes.rate
  disk.write.requests
  disk.write.requests.rate
  image
  image.delete
  image.download
  image.serve
  image.size
  image.update
  image.upload
  instance
  ip.floating
  ip.floating.create
  ip.floating.update
  memory
  memory.resident
  memory.usage
  network.incoming.bytes
  network.incoming.bytes.rate
  network.incoming.packets
  network.incoming.packets.rate
  network.outgoing.bytes
  network.outgoing.bytes.rate
  network.outgoing.packets
  network.outgoing.packets.rate
  network.services.firewall
  network.services.firewall.policy
  network.services.firewall.policy.create
  network.services.firewall.rule
  network.services.firewall.rule.create
  network.services.firewall.rule.update
  network.services.lb.active.connections
  network.services.lb.health_monitor
  network.services.lb.incoming.bytes
  network.services.lb.member
  network.services.lb.member.create
  network.services.lb.member.update
  network.services.lb.outgoing.bytes
  network.services.lb.pool
  network.services.lb.pool.create
  network.services.lb.pool.update
  network.services.lb.total.connections
  network.services.lb.vip
  network.services.lb.vip.create
  network.services.lb.vip.update
  port
  port.create
  port.update
  router
  router.update
  storage.api.request
  storage.containers.objects
  storage.containers.objects.size
  storage.objects
  storage.objects.containers
  storage.objects.incoming.bytes
  storage.objects.outgoing.bytes
  storage.objects.size
  vcpus
*/

func init() {
	flag.Parse()

	parsedLevel, err := log.ParseLevel(*rawLevel)
	if err != nil {
		log.Fatal(err)
	}
	logLevel = parsedLevel
}

var logLevel log.Level = log.InfoLevel
var rawLevel = flag.String("log-level", "info", "log level")
var bindAddr = flag.String("bind-addr", ":9154", "bind address for the metrics server")
var metricsPath = flag.String("metrics-path", "/metrics", "path to metrics endpoint")

func main() {
	log.SetLevel(logLevel)
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

type Scraper struct {
	id         string
	lastScrape time.Time
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

	client, err := openstack.NewTelemetryV2(provider, upstream.EndpointOpts{})
	if err != nil {
		panic(err)
	}

	return &ceilometerCollector{
		metrics: map[string]ceilometerMetric{
			// Hardware metrics
			"cpu": {
				desc: prometheus.NewDesc("openstack_ceilometer_cpu_nanoseconds", "Consumed CPU time (nanoseconds)", []string{"instance_id", "instance_name"}, nil),
				extractLabels: func(sample *meters.OldSample) []string {
					return []string{
						sample.ResourceId,
						sample.ResourceMetadata["display_name"],
					}
				},
			},
			"cpu_util": {
				desc: prometheus.NewDesc("openstack_ceilometer_cpu_percent", "CPU utilization (percent)", []string{"instance_id", "instance_name"}, nil),
				extractLabels: func(sample *meters.OldSample) []string {
					return []string{
						sample.ResourceId,
						sample.ResourceMetadata["display_name"],
					}
				},
			},
			"memory.usage": {
				desc: prometheus.NewDesc("openstack_ceilometer_memory_usage", "Memory utilization", []string{"instance_id", "instance_name"}, nil),
				extractLabels: func(sample *meters.OldSample) []string {
					return []string{
						sample.ResourceId,
						sample.ResourceMetadata["display_name"],
					}
				},
			},
			"memory": {
				desc: prometheus.NewDesc("openstack_ceilometer_memory", "Memory allocation", []string{"instance_id", "instance_name"}, nil),
				extractLabels: func(sample *meters.OldSample) []string {
					return []string{
						sample.ResourceId,
						sample.ResourceMetadata["display_name"],
					}
				},
			},
			"memory.resident": {
				desc: prometheus.NewDesc("openstack_ceilometer_memory_resident", "Resident memory utilization", []string{"instance_id", "instance_name"}, nil),
				extractLabels: func(sample *meters.OldSample) []string {
					return []string{
						sample.ResourceId,
						sample.ResourceMetadata["display_name"],
					}
				},
			},

			// Network
			"network.services.firewall.policy": {
				desc: prometheus.NewDesc("openstack_ceilometer_firewall_policy", "Firewall policy", []string{"name"}, nil),
				extractLabels: func(sample *meters.OldSample) []string {
					return []string{
						sample.ResourceMetadata["name"],
					}
				},
			},
			"network.services.lb.vip": {
				desc: prometheus.NewDesc("openstack_ceilometer_loadbalancer_pool", "Load balancer pool", []string{"name"}, nil),
				extractLabels: func(sample *meters.OldSample) []string {
					return []string{
						sample.ResourceMetadata["name"],
					}
				},
			},
			"network.services.lb.pool": {
				desc: prometheus.NewDesc("openstack_ceilometer_loadbalancer_vip", "Load balancer virtual IP", []string{"name"}, nil),
				extractLabels: func(sample *meters.OldSample) []string {
					return []string{
						sample.ResourceMetadata["name"],
					}
				},
			},
			"network.services.lb.member": {
				desc: prometheus.NewDesc("openstack_ceilometer_loadbalancer_pool_member", "Load balancer pool member", []string{"status", "pool_id"}, nil),
				extractLabels: func(sample *meters.OldSample) []string {
					return []string{
						sample.ResourceMetadata["status"],
						// TODO: Add lookup-table for pool-id -> name
						sample.ResourceMetadata["pool_id"],
					}
				},
			},
			"network.services.lb.incoming.bytes": {
				desc: prometheus.NewDesc("openstack_ceilometer_loadbalancer_pool_bytes_in", "Load balancer pool bytes-in", []string{"pool_id"}, nil),
				extractLabels: func(sample *meters.OldSample) []string {
					return []string{
						// TODO: lookup-table
						sample.ResourceId,
					}
				},
			},
			"network.services.lb.active.connections": {
				desc: prometheus.NewDesc("openstack_ceilometer_loadbalancer_pool_active_connections", "Load balancer pool active connections", []string{"pool_id"}, nil),
				extractLabels: func(sample *meters.OldSample) []string {
					return []string{
						// TODO: lookup-table
						sample.ResourceId,
					}
				},
			},
			// Swift
			"storage.containers.objects": {
				desc: prometheus.NewDesc("openstack_ceilometer_swift_objects", "Swift container objects", []string{"container_id"}, nil),
				extractLabels: func(sample *meters.OldSample) []string {
					return []string{
						strings.SplitN(sample.ResourceId, "/", 2)[1],
					}
				},
			},
			// Usage
			"instance": {
				desc: prometheus.NewDesc("openstack_ceilometer_instance", "Instances", []string{"instance_id", "instance_name", "flavor"}, nil),
				extractLabels: func(sample *meters.OldSample) []string {
					return []string{
						sample.ResourceId,
						sample.ResourceMetadata["display_name"],
						sample.ResourceMetadata["flavor.name"],
					}
				},
			},
		},
		client: client,
	}
}

type ceilometerCollector struct {
	client  *upstream.ServiceClient
	metrics map[string]ceilometerMetric
}
type ceilometerMetric struct {
	desc          *prometheus.Desc
	extractLabels func(*meters.OldSample) []string
}

func (c *ceilometerCollector) Describe(ch chan<- *prometheus.Desc) {
	log.Debugf("Sending %d metrics descriptions", len(c.metrics))
	for _, metric := range c.metrics {
		ch <- metric.desc
	}
}

func (c *ceilometerCollector) Collect(ch chan<- prometheus.Metric) {
	t := time.Now()
	result := make(chan bool)
	defer close(result)
	for resourceLabel, metric := range c.metrics {
		// TODO: Multiple results: ok/notok, time spent, #rows. Labeled per resourceLabel
		go scrape(resourceLabel, metric, c.client, ch, result)
	}
	for _ = range c.metrics {
		<-result
	}

	desc := prometheus.NewDesc("openstack_ceilometer_scrape_duration_ns", "Time taken for scrape", nil, nil)
	ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, float64(time.Since(t).Nanoseconds()))
}

func scrape(resourceLabel string, metric ceilometerMetric, client *upstream.ServiceClient, ch chan<- prometheus.Metric, result chan<- bool) {
	scraper := Scraper{
		id:         "test",
		lastScrape: time.Now().UTC().Add(time.Duration(-5) * time.Minute),
	}

	limit := 200 // TBD
	query := meters.ShowOpts{
		QueryField: "timestamp",
		QueryOp:    "gt",
		QueryValue: scraper.lastScrape.Format("2006-01-02T15:04:05"),
		Limit:      limit,
	}
	log.Debugf("Querying for %v: %v", resourceLabel, query)
	results := meters.Show(client, resourceLabel, query)
	data, err := results.Extract()
	if err != nil {
		log.Warnf("Failed to scrape Ceilometer resource %q for client %v", metric, scraper.id)
		result <- false
		return
	}
	if len(data) == 0 {
		log.Warnf("Query for %v returned no results!", resourceLabel)
		result <- false // TODO: Maybe not to be treated as failure?
		return
	}
	if len(data) == limit {
		log.Warnf("Query for %v returned max number of results (%d), data may be truncated", resourceLabel, limit)
	}
	initialLen := len(data)
	data = deduplicate(data)
	log.Debugf("Query for %s returned %d results, %d remain after deduplication", resourceLabel, initialLen, len(data))

	for _, sample := range data {
		ch <- sampleToMetric(&sample, metric)
	}

	result <- true
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
