#!/bin/bash
#
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

CONTAINER_NAME=chainmaker-docker-vm

docker ps

echo

read -r -p "input container name to stop(default 'chainmaker-docker-vm'): " tmp
if  [ -n "$tmp" ] ;then
  CONTAINER_NAME=$tmp
else
  echo "container name use default: 'chainmaker-docker-vm'"
fi

echo "stop docker vm container"

docker stop "$CONTAINER_NAME"
