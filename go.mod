module chainmaker.org/chainmaker-go

go 1.15

require (
	chainmaker.org/chainmaker-go/accesscontrol v0.0.0
	chainmaker.org/chainmaker-go/blockchain v0.0.0-00010101000000-000000000000
	chainmaker.org/chainmaker-go/consensus v0.0.0
	chainmaker.org/chainmaker-go/net v0.0.0
	chainmaker.org/chainmaker-go/rpcserver v0.0.0
	chainmaker.org/chainmaker-go/txpool v0.0.0
	chainmaker.org/chainmaker-go/vm v0.0.0
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211117131805-630800bfd361
	chainmaker.org/chainmaker/consensus-dpos/v2 v2.0.0-20211119100858-b806ddbb6d35
	chainmaker.org/chainmaker/consensus-hotstuff/v2 v2.0.0-20211119101639-ae06999f7408
	chainmaker.org/chainmaker/consensus-raft/v2 v2.0.0-20211119100947-ce17472c410c
	chainmaker.org/chainmaker/consensus-solo/v2 v2.0.0-20211119095958-4b1cce193b32
	chainmaker.org/chainmaker/consensus-tbft/v2 v2.0.0-20211119101400-96829b485267
	chainmaker.org/chainmaker/consensus-utils/v2 v2.0.0-20211115084213-42e840e1efee
	chainmaker.org/chainmaker/localconf/v2 v2.1.1-0.20211110030026-ce2a7f3760cd
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211117140137-9095a7ab7a69
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211119081550-ff85fecc318d
	chainmaker.org/chainmaker/sdk-go/v2 v2.0.1-0.20211110082824-51f0b56a62ee
	chainmaker.org/chainmaker/txpool-batch/v2 v2.1.1-0.20211109075600-a0a811fe0650
	chainmaker.org/chainmaker/txpool-single/v2 v2.1.1-0.20211109075506-aea78872cdc6
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20211117144316-3f4a940e94f0
	chainmaker.org/chainmaker/vm-evm/v2 v2.1.1-0.20211109084614-e0e363b49d47
	chainmaker.org/chainmaker/vm-gasm/v2 v2.1.1-0.20211109083312-3f36d5d4e4d2
	chainmaker.org/chainmaker/vm-wasmer/v2 v2.1.1-0.20211119040808-92b65899613e
	chainmaker.org/chainmaker/vm-wxvm/v2 v2.1.1-0.20211115073737-27720f7ad7d1
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
	go.opencensus.io v0.23.0 // indirect
	google.golang.org/grpc v1.41.0
)

replace (
	chainmaker.org/chainmaker-go/accesscontrol => ./module/accesscontrol
	chainmaker.org/chainmaker-go/blockchain => ./module/blockchain
	chainmaker.org/chainmaker-go/consensus => ./module/consensus
	chainmaker.org/chainmaker-go/core => ./module/core
	chainmaker.org/chainmaker-go/net => ./module/net
	chainmaker.org/chainmaker-go/rpcserver => ./module/rpcserver
	chainmaker.org/chainmaker-go/snapshot => ./module/snapshot
	chainmaker.org/chainmaker-go/subscriber => ./module/subscriber
	chainmaker.org/chainmaker-go/sync => ./module/sync
	chainmaker.org/chainmaker-go/txpool => ./module/txpool
	chainmaker.org/chainmaker-go/vm => ./module/vm
	github.com/libp2p/go-libp2p-core => chainmaker.org/chainmaker/libp2p-core v0.0.2
	github.com/spf13/afero => github.com/spf13/afero v1.5.1 //for go1.15 build
	github.com/spf13/viper => github.com/spf13/viper v1.7.1 //for go1.15 build
	google.golang.org/grpc v1.40.0 => google.golang.org/grpc v1.26.0
)
