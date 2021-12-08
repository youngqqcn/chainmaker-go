#!/usr/bin/env bash
#
# Copyright (C) BABEC. All rights reserved.
# Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
#
# SPDX-License-Identifier: Apache-2.0
#

set -x
BRANCH=develop
ZXL=develop-zxl
cd ..

go get chainmaker.org/chainmaker/chainconf/v2@${ZXL}
go get chainmaker.org/chainmaker/common/v2@${ZXL}
go get chainmaker.org/chainmaker/consensus/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-chainedbft/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-dpos/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-raft/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-solo/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-tbft/v2@${BRANCH}
go get chainmaker.org/chainmaker/consensus-utils/v2@${BRANCH}
go get chainmaker.org/chainmaker/localconf/v2@${BRANCH}
go get chainmaker.org/chainmaker/logger/v2@${BRANCH}
go get chainmaker.org/chainmaker/net-common@${BRANCH}
go get chainmaker.org/chainmaker/net-libp2p@${BRANCH}
go get chainmaker.org/chainmaker/net-liquid@${BRANCH}
go get chainmaker.org/chainmaker/pb-go/v2@${ZXL}
go get chainmaker.org/chainmaker/protocol/v2@${ZXL}
go get chainmaker.org/chainmaker/sdk-go/v2@${ZXL}
go get chainmaker.org/chainmaker/store/v2@${ZXL}
go get chainmaker.org/chainmaker/txpool-batch/v2@${BRANCH}
go get chainmaker.org/chainmaker/txpool-single/v2@${BRANCH}
go get chainmaker.org/chainmaker/utils/v2@${ZXL}
go get chainmaker.org/chainmaker/vm/v2@${ZXL}
go get chainmaker.org/chainmaker/vm-evm/v2@${ZXL}
go get chainmaker.org/chainmaker/vm-gasm/v2@${ZXL}
go get chainmaker.org/chainmaker/vm-native/v2@${ZXL}
go get chainmaker.org/chainmaker/vm-wasmer/v2@${ZXL}
go get chainmaker.org/chainmaker/vm-wxvm/v2@${BRANCH}
go get chainmaker.org/chainmaker/vm-docker-go@${ZXL}

go mod tidy

cd main
go build -o chainmaker
cd ../tools/cmc
go build -o cmc
