#!/usr/bin/env bash
ENVIRONMENT_UUID=$(curl http://rancher-metadata/latest/self/host/environment_uuid)
/bin/prometheus_pusher -config.customLabels=environmentUUID -config.customLabelValues=${ENVIRONMENT_UUID} $@