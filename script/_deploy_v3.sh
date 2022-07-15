#!/bin/bash

service_root=${1%/}
service=$2


mchs=(root@127.0.0.1)


doDeploy(){
  ssh $1 "rm -rf ${service_root}/${service}.bak"
  ssh $1 "mv ${service_root}/${service} ${service_root}/${service}.bak"
  ssh $1 "systemctl stop ${service}"
  scp ${service} $1:${service_root}/${service}
  ssh $1 "systemctl start ${service}"
}

chmod +x ${service}

for mch in "${mchs[@]}"
do
  doDeploy "$mch"
done
