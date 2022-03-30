#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

# if enable docker vm service and use unix domain socket, run a vm docker container
start_docker_vm() {
  image_name="chainmakerofficial/chainmaker-vm-docker-go:v2.2.1"
  config_file="../config/{org_id}/chainmaker.yml"
  enable_docker_vm=$(grep enable_dockervm $config_file | awk -F: '{gsub(/ /, "", $2);print $2}')
  enable_uds=$(grep uds_open $config_file | awk -F: '{gsub(/ /, "", $2);print $2}')
  if [[ $enable_docker_vm = "true" && $enable_uds = "true" ]]
  then
      mount_path=$(grep dockervm_mount_path $config_file | awk -F: '{gsub(/ /, "", $2);print $2}')
      log_path=$(grep dockervm_log_path $config_file | awk -F: '{gsub(/ /, "", $2);print $2}')
      chain_id=$(grep "chainId:" $config_file | grep -v "#" | awk -F: '{gsub(/ /, "", $2);print $2}')

      if [[ "${mount_path:0:1}" != "/" ]];then
        mount_path=$(pwd)/$mount_path
      fi

      if [[ "${log_path:0:1}" != "/" ]];then
        log_path=$(pwd)/$log_path
      fi

      mkdir -p $mount_path/$chain_id
      mkdir -p $log_path/$chain_id

      echo "run docker vm service container"
      docker run -itd --rm \
        -e ENV_LOG_IN_CONSOLE=true -e ENV_LOG_LEVEL=DEBUG -e ENV_ENABLE_UDS=true \
        -v $mount_path/$chain_id:/mount \
        -v $log_path/$chain_id:/log \
        --name DOCKERVM-{org_id}-$chain_id \
        --privileged $image_name

      retval="$?"
      if [ $retval -ne 0 ]
      then
        echo "trying to remove existing container"
        docker stop DOCKERVM-{org_id}-$chain_id
        docker rm DOCKERVM-{org_id}-$chain_id
        docker run -itd --rm \
          -e ENV_LOG_IN_CONSOLE=true -e ENV_LOG_LEVEL=DEBUG -e ENV_ENABLE_UDS=true \
          -v $mount_path/$chain_id:/mount \
          -v $log_path/$chain_id:/log \
          --name DOCKERVM-{org_id}-$chain_id \
          --privileged $image_name
      fi

  fi
}

export LD_LIBRARY_PATH=$(dirname $PWD)/lib:$LD_LIBRARY_PATH
export PATH=$(dirname $PWD)/lib:$PATH
export WASMER_BACKTRACE=1
pid=`ps -ef | grep chainmaker | grep "\-c ../config/{org_id}/chainmaker.yml" | grep -v grep |  awk  '{print $2}'`
if [ -z ${pid} ];then
    # check if need to start docker vm service.
    start_docker_vm
    #nohup ./chainmaker start -c ../config/{org_id}/chainmaker.yml > /dev/null 2>&1 &
    nohup ./chainmaker start -c ../config/{org_id}/chainmaker.yml > panic.log 2>&1 &
    echo "chainmaker is startting, pls check log..."
else
    echo "chainmaker is already started"
fi

