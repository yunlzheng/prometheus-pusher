# prometheus-pusher

collection prometheus data and push to pushgateway

## Architecture

![](http://7pn5d3.com1.z0.glb.clouddn.com/prometheus_pusher.png)

## Feature

* Support Prometheus Config File
* Support Prometheus Service Discovery
* Support Custom Label

## How To Use

Prepare prometheus config file **/ect/prom-conf/prometheus.yml**

```
global:                                                                         
  scrape_interval:     15s
  evaluation_interval: 15s
  external_labels:                                                              
      monitor: 'exporter-metrics'                                                
                                                                                
scrape_configs:

- job_name: 'HostsMetrics'
  dns_sd_configs:
  - names:
    - node-exporter
    refresh_interval: 15s
    type: A
    port: 9100

- job_name: 'ContainerMetrics'
  static_configs:
    - targets:
      - 'rancher-server:9108'
- job_name: 'RancherServerMetrics'
  dns_sd_configs:
  - names:
    - cadvisor
    refresh_interval: 15s
    type: A
    port: 8080

- job_name: 'RancherApi'
  dns_sd_configs:
  - names:
    - 'prometheus-rancher-exporter'
    refresh_interval: 15s
    type: A
    port: 9173

- job_name: 'Prometheus'
  static_configs:
    - targets:
      - '127.0.0.1:9090'

```

In command line:

```
export PUSH_GATEWAY=http://pushgateway.example.org:9091
./prometheus_pusher -config.file=prometheus.yml 
```

In docker-compose

> Note: you should set the environment variable of PushGateway address 

```
version: '2'
services:
  pusher:
    image: wisecity/prometheus-pusher
    environment:
      PUSH_GATEWAY: http://pushgateway.example.org:9091
    volumes:
    - /ect/prom-conf:/etc/prom-conf
    entrypoint:
    - /bin/prometheus_pusher
    - -config.file
    - /etc/prom-conf/prometheus.yml
```

### Add custom metrics labels

In some case, if you want add external metrics key for the origin metrics data. You can use customLabels.

In our case, we collection container data from mutil rancher environment with cadvistor. 
We want the prometheus query express can precise positioning the container from different environment. So we add rancher environment uuid as the custom label.

> Note. customLabel will overwrite the origin metrics value

```
./prometheus_pusher -config.file=prometheus.yml -config.customLabels=label1,label2 -config.customLabelValues=value1,value2
```
