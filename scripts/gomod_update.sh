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
  BRANCH="develop"
fi
QC="v2.2.0_alpha_qc"
cd ..

go get chainmaker.org/chainmaker/chainconf/v2@${QC}
go get chainmaker.org/chainmaker/common/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-maxbft/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-dpos/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-raft/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-solo/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-tbft/v2@${QC}
go get chainmaker.org/chainmaker/consensus-utils/v2@${QC}
go get chainmaker.org/chainmaker/localconf/v2@${QC}
go get chainmaker.org/chainmaker/logger/v2@${BRANCH}
go get chainmaker.org/chainmaker/net-common@${BRANCH}
go get chainmaker.org/chainmaker/net-libp2p@${BRANCH}
go get chainmaker.org/chainmaker/net-liquid@${BRANCH}
go get chainmaker.org/chainmaker/pb-go/v2@${QC}
go get chainmaker.org/chainmaker/protocol/v2@${QC}
go get chainmaker.org/chainmaker/sdk-go/v2@${QC}
go get chainmaker.org/chainmaker/store/v2@${QC}
go get chainmaker.org/chainmaker/txpool-batch/v2@${QC}
go get chainmaker.org/chainmaker/txpool-single/v2@${QC}
go get chainmaker.org/chainmaker/utils/v2@${QC}
go get chainmaker.org/chainmaker/vm-docker-go/v2@${QC}
go get chainmaker.org/chainmaker/vm-evm/v2@${BRANCH}
go get chainmaker.org/chainmaker/vm-gasm/v2@${BRANCH}
go get chainmaker.org/chainmaker/vm-native/v2@${QC}
go get chainmaker.org/chainmaker/vm-wasmer/v2@${BRANCH}
go get chainmaker.org/chainmaker/vm-wxvm/v2@${BRANCH}
go get chainmaker.org/chainmaker/vm/v2@${QC}

go mod tidy

make
make cmc