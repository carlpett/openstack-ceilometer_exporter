# Openstack Ceilometer exporter for Prometheus
This is a [Prometheus](prometheus.io) exporter that scrapes the Openstack Ceilometer API and exposes it as Prometheus metrics.

# Running
`openstack_ceilometer_exporter [flags]`

## Flags
| Name              | Description                                                    | Default  |
|-------------------|----------------------------------------------------------------|----------|
| -bind-addr        | bind address for the metrics server                            | :9181    |
| -metrics-path     | path to metrics endpoint                                       | /metrics |
| -disabled-metrics | comma-separated list of metrics to disable (supports globbing) |          |
| -enabled-metrics  | comma-separated list of metrics to enable (supports globbing)  | *        |
| -max-metric-age   | maximum age of metrics to retrieve                             | 5m       |
| -max-results      | maximum number of results to fetch for any metric              | 100      |
| -help             | shows help                                                     |          |
| -list-metrics     | list available metrics and exit                                |          |

# Building
Just `go build`!
