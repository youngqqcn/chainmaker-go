module chainmaker.org/chainmaker-go

go 1.15

require (
	chainmaker.org/chainmaker-go/accesscontrol v0.0.0
	chainmaker.org/chainmaker-go/blockchain v0.0.0-00010101000000-000000000000
	chainmaker.org/chainmaker-go/core v0.0.0-00010101000000-000000000000 // indirect
	chainmaker.org/chainmaker-go/net v0.0.0
	chainmaker.org/chainmaker-go/rpcserver v0.0.0-00010101000000-000000000000
	chainmaker.org/chainmaker-go/txpool v0.0.0
	chainmaker.org/chainmaker-go/vm v0.0.0
	chainmaker.org/chainmaker/common/v2 v2.1.0
	chainmaker.org/chainmaker/consensus-chainedbft/v2 v2.0.0-20211112163433-f0dab0eb58b8
	chainmaker.org/chainmaker/consensus-dpos/v2 v2.0.0-20211112092735-d15ea84c5f44
	chainmaker.org/chainmaker/consensus-raft/v2 v2.0.0-20211112163240-bd1b63cc16bb
	chainmaker.org/chainmaker/consensus-solo/v2 v2.0.0-20211112091638-d0d658ddbdfa
	chainmaker.org/chainmaker/consensus-tbft/v2 v2.0.0-20211112163347-05006087b5d8
	chainmaker.org/chainmaker/consensus-utils/v2 v2.0.0-20211112075449-b2b8e6366203
	chainmaker.org/chainmaker/consensus/v2 v2.0.0-20211112163732-66772d402b44
	chainmaker.org/chainmaker/localconf/v2 v2.0.0-20211101111610-0d268248b5c8
	chainmaker.org/chainmaker/logger/v2 v2.1.0
	chainmaker.org/chainmaker/pb-go/v2 v2.1.0
	chainmaker.org/chainmaker/protocol/v2 v2.0.1-0.20211108075639-576c31f03396
	chainmaker.org/chainmaker/sdk-go/v2 v2.1.0
	chainmaker.org/chainmaker/txpool-batch/v2 v2.0.0-20211019074609-46e3d29f0908
	chainmaker.org/chainmaker/txpool-single/v2 v2.0.0-20211018131403-7eb37f80a128
	chainmaker.org/chainmaker/utils/v2 v2.0.0-20211108092352-2a3335a4ba15
	chainmaker.org/chainmaker/vm-evm v0.0.0-20211015132845-e5b020e52194
	chainmaker.org/chainmaker/vm-gasm v0.0.0-20211101123646-aed5e0b2eeed
	chainmaker.org/chainmaker/vm-wasmer v0.0.0-20211102025640-44ec33122e8c
	chainmaker.org/chainmaker/vm-wxvm v0.0.0-20211015133128-53c7b2ac262f
	code.cloudfoundry.org/bytefmt v0.0.0-20211005130812-5bb3c17173e5
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/ethereum/go-ethereum v1.10.4
	github.com/gogo/protobuf v1.3.2
	github.com/mr-tron/base58 v1.2.0
	github.com/prometheus/client_golang v1.11.0
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc v1.41.0
)

replace (
	chainmaker.org/chainmaker-go/accesscontrol => ./module/accesscontrol
	chainmaker.org/chainmaker-go/blockchain => ./module/blockchain
	chainmaker.org/chainmaker-go/core => ./module/core
	chainmaker.org/chainmaker-go/net => ./module/net
	chainmaker.org/chainmaker-go/rpcserver => ./module/rpcserver
	chainmaker.org/chainmaker-go/snapshot => ./module/snapshot
	chainmaker.org/chainmaker-go/subscriber => ./module/subscriber
	chainmaker.org/chainmaker-go/sync => ./module/sync
	chainmaker.org/chainmaker-go/txpool => ./module/txpool
	chainmaker.org/chainmaker-go/vm => ./module/vm
	github.com/libp2p/go-libp2p-core => chainmaker.org/chainmaker/libp2p-core v0.0.2
	google.golang.org/grpc v1.40.0 => google.golang.org/grpc v1.26.0
)
