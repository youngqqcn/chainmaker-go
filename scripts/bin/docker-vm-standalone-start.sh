#!/bin/bash
#
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

LOG_PATH=$(pwd)/log
MOUNT_PATH=$(pwd)/docker-go
LOG_LEVEL=INFO
EXPOSE_PORT=22351
CONTAINER_NAME=chainmaker-docker-vm
IMAGE_NAME="chainmakerofficial/chainmaker-vm-docker-go:v2.2.1"


read -r -p "input path to cache contract files(must be absolute path, default:'./docker-go'): " tmp
if  [ -n "$tmp" ] ;then
  MOUNT_PATH=$tmp
fi

if  [ ! -d "$MOUNT_PATH" ];then
  read -r -p "contracts path does not exist, create it or not(y|n): " need_create
  if [ "$need_create" == "yes" ] || [ "$need_create" == "y" ]; then
    mkdir -p "$MOUNT_PATH"
    if [ $? -ne 0 ]; then
      echo "create contracts path failed. exit"
      exit 1
    fi
  else
    exit 1
  fi
fi


read -r -p "input log path(must be absolute path, default:'./log'): " tmp
if  [ -n "$tmp" ] ;then
  LOG_PATH=$tmp
fi

if  [ ! -d "$LOG_PATH" ];then
  read -r -p "log path does not exist, create it or not(y|n): " need_create
  if [ "$need_create" == "yes" ] || [ "$need_create" == "y" ]; then
    mkdir -p "$LOG_PATH"
    if [ $? -ne 0 ]; then
      echo "create log path failed. exit"
      exit 1
    fi
  else
    exit 1
  fi
fi

read -r -p "input log level(DEBUG|INFO(default)|WARN|ERROR): " tmp
if  [ -n "$tmp" ] ;then
  if  [ $tmp == "DEBUG" ] || [ $tmp == "INFO" ] || [ $tmp == "WARN" ] || [ $tmp == "ERROR" ];then
      LOG_LEVEL=$tmp
  else
    echo "unknown log level [" $tmp "], so use default"
  fi
fi

read -r -p "input expose port(default 22351): " tmp
if  [ -n "$tmp" ] ;then
  if [[ $tmp =~ ^[0-9]+$ ]] ;then
      EXPOSE_PORT=$tmp
  else
    echo "unknown expose port [" $tmp "], so use 22351"
  fi
fi


read -r -p "input container name(default 'chainmaker-docker-vm'): " tmp
if  [ -n "$tmp" ] ;then
  CONTAINER_NAME=$tmp
else
  echo "container name use default: 'chainmaker-docker-vm'"
fi


exist=$(docker ps -a -f name="$CONTAINER_NAME" --format '{{.Names}}')
if [ "$exist" ]; then
  echo "container is exist, please remove container first."
  exit 1
fi

echo "start docker vm container"
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
docker run -itd --rm \
  -e ENV_LOG_IN_CONSOLE=false -e ENV_LOG_LEVEL="$LOG_LEVEL" \
  -e ENV_USER_NUM=1000 -e ENV_MAX_CONCURRENCY=100 -e ENV_TX_TIME_LIMIT=8 \
  -v "$LOG_PATH":/log \
  -p "$EXPOSE_PORT":22359 \
  --name "$CONTAINER_NAME" \
  --privileged $IMAGE_NAME

docker ps -a -f name="$CONTAINER_NAME"