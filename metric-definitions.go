package main

import (
	"fmt"
	"strings"

	"github.com/DSpeichert/gophercloud/openstack/telemetry/v2/meters"

	"github.com/prometheus/client_golang/prometheus"
)

func getMetrics(lookupSvc *LookupService) *map[string]ceilometerMetric {
	return &map[string]ceilometerMetric{
		// Hardware metrics
		"cpu": {
			desc: prometheus.NewDesc(makeFQName("cpu_nanoseconds"), "Consumed CPU time (nanoseconds)", []string{"instance_id", "instance_name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
				}
			},
		},
		"cpu_util": {
			desc: prometheus.NewDesc(makeFQName("cpu_percent"), "CPU utilization (percent)", []string{"instance_id", "instance_name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
				}
			},
		},
		"disk.allocation": {
			desc: prometheus.NewDesc(makeFQName("disk_allocation"), "Disk allocation", []string{"instance_id", "instance_name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
				}
			},
		},
		"disk.capacity": {
			desc: prometheus.NewDesc(makeFQName("disk_capacity"), "Disk capacity", []string{"instance_id", "instance_name", "device"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
					sample.ResourceMetadata["device"],
				}
			},
		},
		"disk.ephemeral.size": {
			desc: prometheus.NewDesc(makeFQName("disk_ephemeral_size"), "Size of ephemeral disk  ", []string{"instance_id", "instance_name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
				}
			},
		},
		"disk.read.bytes": {
			desc: prometheus.NewDesc(makeFQName("disk_read_bytes"), "Disk bytes read", []string{"instance_id", "instance_name", "device"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
					sample.ResourceMetadata["device"],
				}
			},
		},
		"disk.read.requests": {
			desc: prometheus.NewDesc(makeFQName("disk_read_requests"), "Disk read requests", []string{"instance_id", "instance_name", "device"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
					sample.ResourceMetadata["device"],
				}
			},
		},
		"disk.root.size": {
			desc: prometheus.NewDesc(makeFQName("disk_root_size"), "Root disk size", []string{"instance_id", "instance_name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
				}
			},
		},
		"disk.usage": {
			desc: prometheus.NewDesc(makeFQName("disk_usage"), "Disk usage", []string{"instance_id", "instance_name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
				}
			},
		},
		"disk.write.bytes": {
			desc: prometheus.NewDesc(makeFQName("disk_write_bytes"), "Disk written bytes", []string{"instance_id", "instance_name", "device"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
					sample.ResourceMetadata["device"],
				}
			},
		},
		"disk.write.requests": {
			desc: prometheus.NewDesc(makeFQName("disk_write_requests"), "Disk write requests", []string{"instance_id", "instance_name", "device"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
					sample.ResourceMetadata["device"],
				}
			},
		},

		"memory.usage": {
			desc: prometheus.NewDesc(makeFQName("memory_usage"), "Memory utilization", []string{"instance_id", "instance_name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
				}
			},
		},
		"memory": {
			desc: prometheus.NewDesc(makeFQName("memory"), "Memory allocation", []string{"instance_id", "instance_name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
				}
			},
		},
		"memory.resident": {
			desc: prometheus.NewDesc(makeFQName("memory_resident"), "Resident memory utilization", []string{"instance_id", "instance_name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
				}
			},
		},
		"network.incoming.bytes": {
			desc: prometheus.NewDesc(makeFQName("incoming_bytes"), "Instance incoming network (bytes)", []string{"instance_id", "instance_name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceMetadata["instance_id"],
					lookupSvc.lookupInstance(sample.ResourceMetadata["instance_id"]),
				}
			},
		},
		"network.incoming.packets": {
			desc: prometheus.NewDesc(makeFQName("incoming_packets"), "Instance incoming network (packets)", []string{"instance_id", "instance_name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceMetadata["instance_id"],
					lookupSvc.lookupInstance(sample.ResourceMetadata["instance_id"]),
				}
			},
		},
		"network.outgoing.bytes": {
			desc: prometheus.NewDesc(makeFQName("outgoing_bytes"), "Instance outgoing network (bytes)", []string{"instance_id", "instance_name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceMetadata["instance_id"],
					lookupSvc.lookupInstance(sample.ResourceMetadata["instance_id"]),
				}
			},
		},
		"network.outgoing.packets": {
			desc: prometheus.NewDesc(makeFQName("outgoing_packets"), "Instance outgoing network (packets)", []string{"instance_id", "instance_name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceMetadata["instance_id"],
					lookupSvc.lookupInstance(sample.ResourceMetadata["instance_id"]),
				}
			},
		},
		// Network
		"network.services.firewall.policy": {
			desc: prometheus.NewDesc(makeFQName("firewall_policy"), "Firewall policy", []string{"name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceMetadata["name"],
				}
			},
		},
		"network.services.lb.vip": {
			desc: prometheus.NewDesc(makeFQName("loadbalancer_pool"), "Load balancer pool", []string{"name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceMetadata["name"],
				}
			},
		},
		"network.services.lb.pool": {
			desc: prometheus.NewDesc(makeFQName("loadbalancer_vip"), "Load balancer virtual IP", []string{"name"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceMetadata["name"],
				}
			},
		},
		"network.services.lb.member": {
			desc: prometheus.NewDesc(makeFQName("loadbalancer_pool_member"), "Load balancer pool member", []string{"member", "status", "pool"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					fmt.Sprintf("%s:%s", sample.ResourceMetadata["address"], sample.ResourceMetadata["protocol_port"]),
					sample.ResourceMetadata["status"],
					lookupSvc.lookupPool(sample.ResourceMetadata["pool_id"]),
				}
			},
		},
		"network.services.lb.incoming.bytes": {
			desc: prometheus.NewDesc(makeFQName("loadbalancer_pool_bytes_in"), "Load balancer pool bytes-in", []string{"pool"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					lookupSvc.lookupPool(sample.ResourceId),
				}
			},
		},
		"network.services.lb.outgoing.bytes": {
			desc: prometheus.NewDesc(makeFQName("loadbalancer_pool_bytes_out"), "Load balancer pool bytes-out", []string{"pool"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					lookupSvc.lookupPool(sample.ResourceId),
				}
			},
		},
		"network.services.lb.active.connections": {
			desc: prometheus.NewDesc(makeFQName("loadbalancer_pool_active_connections"), "Load balancer pool active connections", []string{"pool"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					lookupSvc.lookupPool(sample.ResourceId),
				}
			},
		},
		"network.services.lb.total.connections": {
			desc: prometheus.NewDesc(makeFQName("loadbalancer_pool_total_connections"), "Load balancer pool total connections", []string{"pool"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					lookupSvc.lookupPool(sample.ResourceId),
				}
			},
		},
		// Swift
		"storage.containers.objects": {
			desc: prometheus.NewDesc(makeFQName("swift_objects"), "Swift container objects", []string{"container_id"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					strings.SplitN(sample.ResourceId, "/", 2)[1],
				}
			},
		},
		"storage.containers.objects.size": {
			desc: prometheus.NewDesc(makeFQName("swift_objects_size"), "Swift container size (bytes)", []string{"container_id"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					strings.SplitN(sample.ResourceId, "/", 2)[1],
				}
			},
		},
		// Usage
		"instance": {
			desc: prometheus.NewDesc(makeFQName("instance"), "Instances", []string{"instance_id", "instance_name", "flavor"}, nil),
			extractLabels: func(sample *meters.OldSample) []string {
				return []string{
					sample.ResourceId,
					sample.ResourceMetadata["display_name"],
					sample.ResourceMetadata["flavor.name"],
				}
			},
		},
	}
}
