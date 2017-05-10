#!/usr/bin/env bash

ACTIVE_PROFILE=${PROFILE:=default}
echo 'ACTIVE PROFILE :'${ACTIVE_PROFILE}

case ${ACTIVE_PROFILE} in
'rancher')

/bin/prometheus_pusher -config.customLabels=environmentUUID -config.customLabelValues=$(curl http://rancher-metadata/latest/self/host/environment_uuid) $@

;;
*)

/bin/prometheus_pusher $@

esac

