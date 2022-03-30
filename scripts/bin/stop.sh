#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

pid=`ps -ef | grep chainmaker | grep "\-c ../config/{org_id}/chainmaker.yml" | grep -v grep |  awk  '{print $2}'`
if [ ! -z ${pid} ];then
    kill $pid
fi

# if enable docker vm service and use unix domain socket, stop the running container
stop_docker_vm() {
  config_file="../config/{org_id}/chainmaker.yml"
  enable_docker_vm=$(grep enable_dockervm $config_file | awk -F: '{gsub(/ /, "", $2);print $2}')
  enable_uds=$(grep uds_open $config_file | awk -F: '{gsub(/ /, "", $2);print $2}')
  chain_id=$(grep "chainId:" $config_file | grep -v "#" | awk -F: '{gsub(/ /, "", $2);print $2}')
  container_name=DOCKERVM-{org_id}-$chain_id
  if [[ $enable_docker_vm = "true" && $enable_uds = "true" ]]
  then
    container_exists=$(docker ps -f name="$container_name" --format '{{.Names}}')
    if [[ $container_exists ]]; then
      echo "stop docker vm container:"
      docker stop "$container_name"
    fi
  fi
}
stop_docker_vm

echo "chainmaker is stopped"
