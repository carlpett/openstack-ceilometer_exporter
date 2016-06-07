Metrics not yet implemented

  disk.device.allocation                    disk.device.* seems to give same output as disk.*
  disk.device.capacity                      disk.device.* seems to give same output as disk.*
  disk.device.read.bytes                    disk.device.* seems to give same output as disk.*
  disk.device.read.bytes.rate               Pre-aggregated
  disk.device.read.requests                 disk.device.* seems to give same output as disk.*
  disk.device.read.requests.rate            Pre-aggregated
  disk.device.usage                         disk.device.* seems to give same output as disk.*
  disk.device.write.bytes                   disk.device.* seems to give same output as disk.*
  disk.device.write.bytes.rate              Pre-aggregated
  disk.device.write.requests                disk.device.* seems to give same output as disk.*
  disk.device.write.requests.rate           Pre-aggregated
  disk.read.bytes.rate                      Pre-aggregated
  disk.read.requests.rate                   Pre-aggregated
  disk.write.bytes.rate                     Pre-aggregated
  disk.write.requests.rate                  Pre-aggregated
  image                                     ?
  image.delete                              Events
  image.download                            Events
  image.serve                               Events
  image.size                                Events
  image.update                              Events
  image.upload                              Events
  network.incoming.bytes.rate               Pre-aggregated, ignore
  network.incoming.packets.rate             Pre-aggregated, ignore
  network.outgoing.bytes.rate               Pre-aggregated, ignore
  network.outgoing.packets.rate             Pre-aggregated, ignore
  network.services.firewall                 Active/inactive and admin_state - not very interesting
  network.services.firewall.rule            Rule details - not interesting
  network.services.lb.health_monitor        Shows connected pools - interesting?
  port                                      Mainly port events - no stats
  router                                    Events -"-
  storage.api.request                       Needs to be aggregated, shows every request to every object!
  storage.objects.incoming.bytes            -"-
  storage.objects.outgoing.bytes            -"-
  storage.objects                           Aggregated view of all containers
  storage.objects.size                      -"-
  vcpus                                     Times out server?
