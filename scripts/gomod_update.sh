#!/usr/bin/env bash
#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -x
BRANCH=$1
if [[ ! -n $BRANCH ]]; then
  BRANCH="v2.2.0"
fi
NET_VERSION="v1.1.0"
V221="v2.2.1_qc"
cd ..

go get chainmaker.org/chainmaker/chainconf/v2@${V221}
go get chainmaker.org/chainmaker/common/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-maxbft/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-dpos/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-raft/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-solo/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-tbft/v2@${V221}
go get chainmaker.org/chainmaker/consensus-utils/v2@${BRANCH}
go get chainmaker.org/chainmaker/localconf/v2@${BRANCH}
go get chainmaker.org/chainmaker/logger/v2@${BRANCH}
go get chainmaker.org/chainmaker/net-common@${NET_VERSION}
go get chainmaker.org/chainmaker/net-libp2p@v1.1.1_qc
go get chainmaker.org/chainmaker/net-liquid@v1.0.0
go get chainmaker.org/chainmaker/pb-go/v2@${V221}
go get chainmaker.org/chainmaker/protocol/v2@${V221}
go get chainmaker.org/chainmaker/sdk-go/v2@${V221}
go get chainmaker.org/chainmaker/store/v2@${V221}
go get chainmaker.org/chainmaker/txpool-batch/v2@${V221}
go get chainmaker.org/chainmaker/txpool-single/v2@${V221}
go get chainmaker.org/chainmaker/utils/v2@${BRANCH}
go get chainmaker.org/chainmaker/vm-docker-go/v2@${V221}
go get chainmaker.org/chainmaker/vm-evm/v2@${BRANCH}
go get chainmaker.org/chainmaker/vm-gasm/v2@${BRANCH}
go get chainmaker.org/chainmaker/vm-native/v2@${BRANCH}
go get chainmaker.org/chainmaker/vm-wasmer/v2@${BRANCH}
go get chainmaker.org/chainmaker/vm-wxvm/v2@${BRANCH}
go get chainmaker.org/chainmaker/vm/v2@${BRANCH}

go mod tidy

make
make cmc