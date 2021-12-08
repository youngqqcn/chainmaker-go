module chainmaker.org/chainmaker-go

go 1.15

require (
	chainmaker.org/chainmaker/chainconf/v2 v2.1.1-0.20211201032240-900b959567cf
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211207133409-8dbaf5ac3afc
	chainmaker.org/chainmaker/consensus-chainedbft/v2 v2.0.0-20211207134138-c082d96d81e8
	chainmaker.org/chainmaker/consensus-dpos/v2 v2.0.0-20211203091348-0198703a6653
	chainmaker.org/chainmaker/consensus-raft/v2 v2.0.0-20211208030426-88ad294d0250
	chainmaker.org/chainmaker/consensus-solo/v2 v2.0.0-20211123123020-d0f19f91f625
	chainmaker.org/chainmaker/consensus-tbft/v2 v2.0.0-20211207123754-393bcafda1cb
	chainmaker.org/chainmaker/consensus-utils/v2 v2.0.0-20211207122912-c758e2108ff7
	chainmaker.org/chainmaker/consensus/v2 v2.0.0-20211203091717-867d13f316b0
	chainmaker.org/chainmaker/localconf/v2 v2.1.1-0.20211110030026-ce2a7f3760cd
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/net-common v0.0.7-0.20211109085844-739f0f904b96
	chainmaker.org/chainmaker/net-libp2p v1.0.1-0.20211109090515-4889a63c74af
	chainmaker.org/chainmaker/net-liquid v1.0.1-0.20211122114338-22ed0765724f
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211207133220-7a8e48b1655a
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211207133449-dcd91d474c06
	chainmaker.org/chainmaker/sdk-go/v2 v2.0.1-0.20211207024428-c8454054fd8a
	chainmaker.org/chainmaker/store-sqldb/v2 v2.1.1-0.20211202031309-159eb01a87b1 // indirect
	chainmaker.org/chainmaker/store/v2 v2.1.1-0.20211122063451-8a41802784e4
	chainmaker.org/chainmaker/txpool-batch/v2 v2.1.1-0.20211129022941-e7a476018d0c
	chainmaker.org/chainmaker/txpool-single/v2 v2.1.1-0.20211109075506-aea78872cdc6
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20211208040318-ad58981b7d09
	chainmaker.org/chainmaker/vm-docker-go v0.0.0-20211207085346-9fd71eabcc42
	chainmaker.org/chainmaker/vm-evm/v2 v2.0.0-20211130025150-83cc8c18e71b
	chainmaker.org/chainmaker/vm-gasm/v2 v2.1.1-0.20211201082727-f30e454c201e
	chainmaker.org/chainmaker/vm-native/v2 v2.1.1-0.20211208015549-a596967e8ef9
	chainmaker.org/chainmaker/vm-wasmer/v2 v2.1.1-0.20211129130826-b242ad73cae2
	chainmaker.org/chainmaker/vm/v2 v2.1.1-0.20211207134238-dc2563d0c03f
	code.cloudfoundry.org/bytefmt v0.0.0-20211005130812-5bb3c17173e5
	github.com/Rican7/retry v0.1.0
	github.com/Workiva/go-datastructures v1.0.53
	github.com/c-bata/go-prompt v0.2.2
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/ethereum/go-ethereum v1.10.4
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/google/shlex v0.0.0-20181106134648-c34317bd91bf
	github.com/gosuri/uilive v0.0.4 // indirect
	github.com/gosuri/uiprogress v0.0.1
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/hokaccha/go-prettyjson v0.0.0-20201222001619-a42f9ac2ec8e
	github.com/holiman/uint256 v1.2.0
	github.com/hpcloud/tail v1.0.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/mr-tron/base58 v1.2.0
	github.com/panjf2000/ants/v2 v2.4.6
	github.com/prometheus/client_golang v1.11.0
	github.com/rcrowley/go-metrics v0.0.0-20190826022208-cac0b30c2563
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	github.com/syndtr/goleveldb v1.0.1-0.20210305035536-64b5b1c73954
	github.com/tidwall/pretty v1.0.2
	go.opencensus.io v0.23.0 // indirect
	go.uber.org/atomic v1.7.0
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/time v0.0.0-20210608053304-ed9ce3a009e4
	google.golang.org/grpc v1.41.0
	gorm.io/driver/mysql v1.2.0
	gorm.io/gorm v1.22.3
)

replace (
	github.com/libp2p/go-libp2p-core => chainmaker.org/chainmaker/libp2p-core v0.0.2
	github.com/spf13/afero => github.com/spf13/afero v1.5.1 //for go1.15 build
	github.com/spf13/viper => github.com/spf13/viper v1.7.1 //for go1.15 build
	google.golang.org/grpc v1.40.0 => google.golang.org/grpc v1.26.0
)
