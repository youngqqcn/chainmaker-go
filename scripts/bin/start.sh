#!/bin/bash
#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

export LD_LIBRARY_PATH=$(dirname $PWD)/lib:$LD_LIBRARY_PATH
export PATH=$(dirname $PWD)/lib:$PATH
export WASMER_BACKTRACE=1

config_file="../config/{org_id}/chainmaker.yml"

# if clean existed container(can be -y/-f/force)
FORCE_CLEAN=$1

# if enable docker vm service and use unix domain socket, run a vm docker container
start_docker_vm() {
  image_name="chainmakerofficial/chainmaker-vm-docker-go:v2.2.1"

  container_name=DOCKERVM-{org_id}
  echo "start docker vm service container: $container_name"
  #check container exists
  exist=$(docker ps -f name="$container_name" --format '{{.Names}}')
  if [ "$exist" ]; then
    echo "$container_name already RUNNING, please stop it first."
    exit 1
  fi

  exist=$(docker ps -a -f name="$container_name" --format '{{.Names}}')
  if [ "$exist" ]; then
    echo "$container_name already exists(STOPPED)"
    if [[ "$FORCE_CLEAN" == "-f" ]] || [ "$FORCE_CLEAN" == "force" ] || [ "$FORCE_CLEAN" == "-y" ]; then
      echo "remove it:"
      docker rm $container_name
    else
      read -r -p "remove it and start a new container, default: yes (y|n): " need_rm
      if [ "$need_rm" == "no" ] || [ "$need_rm" == "n" ]; then
        exit 0
      else
        docker rm $container_name
      fi
    fi
  fi

  # concat mount_path and log_path for container to mount
  mount_path=$(grep dockervm_mount_path $config_file | awk -F: '{gsub(/ /, "", $2);print $2}')
  log_path=$(grep dockervm_log_path $config_file | awk -F: '{gsub(/ /, "", $2);print $2}')
  log_level=$(grep log_level $config_file | awk -F: '{gsub(/ /, "", $2);print $2}')
  log_in_console=$(grep log_in_console $config_file | awk -F: '{gsub(/ /, "", $2);print $2}')
  if [[ "${mount_path:0:1}" != "/" ]];then
    mount_path=$(pwd)/$mount_path
  fi
  if [[ "${log_path:0:1}" != "/" ]];then
    log_path=$(pwd)/$log_path
  fi

  mkdir -p "$mount_path"
  mkdir -p "$log_path"

  # env params:
  # ENV_ENABLE_UDS=false
  # ENV_USER_NUM=100
  # ENV_TX_TIME_LIMIT=2
  # ENV_LOG_LEVEL=INFO
  # ENV_LOG_IN_CONSOLE=false
  # ENV_MAX_CONCURRENCY=50
  # ENV_VM_SERVICE_PORT=22359
  # ENV_ENABLE_PPROF=
  # ENV_PPROF_PORT=
  echo "start docker vm service container:"
  docker run -itd \
    -e ENV_LOG_IN_CONSOLE="$log_in_console" -e ENV_LOG_LEVEL="$log_level" -e ENV_ENABLE_UDS=true \
    -e ENV_USER_NUM=1000 -e ENV_MAX_CONCURRENCY=100 -e ENV_TX_TIME_LIMIT=8 \
    -v "$mount_path":/mount \
    -v "$log_path":/log \
    --name DOCKERVM-{org_id} \
    --privileged $image_name

  retval="$?"
  if [ $retval -ne 0 ]; then
    echo "Fail to run docker vm."
    exit 1
  fi

  echo "waiting for docker vm container to warm up..."
  sleep 5
}


pid=`ps -ef | grep chainmaker | grep "\-c ../config/{org_id}/chainmaker.yml" | grep -v grep |  awk  '{print $2}'`
if [ -z ${pid} ];then

    # check if need to start docker vm service.
    enable_docker_vm=$(grep enable_dockervm $config_file | awk -F: '{gsub(/ /, "", $2);print $2}')
    enable_uds=$(grep uds_open $config_file | awk -F: '{gsub(/ /, "", $2);print $2}')
    if [[ $enable_docker_vm = "true" && $enable_uds = "true" ]]
    then
      start_docker_vm
    fi

    # start chainmaker
    #nohup ./chainmaker start -c ../config/{org_id}/chainmaker.yml > /dev/null 2>&1 &
    nohup ./chainmaker start -c ../config/{org_id}/chainmaker.yml > panic.log 2>&1 &
    echo "chainmaker is startting, pls check log..."
else
    echo "chainmaker is already started"
fi

