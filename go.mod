module chainmaker.org/chainmaker-go

go 1.15

require (
	chainmaker.org/chainmaker/chainconf/v2 v2.1.1-0.20211209080919-d14a1d8422bd
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211207133409-8dbaf5ac3afc
	chainmaker.org/chainmaker/consensus-chainedbft/v2 v2.0.0-20211207134138-c082d96d81e8
	chainmaker.org/chainmaker/consensus-dpos/v2 v2.0.0-20211210081306-4784edc9d839
	chainmaker.org/chainmaker/consensus-maxbft/v2 v2.0.0-20211210081411-5ef693cd806f
	chainmaker.org/chainmaker/consensus-raft/v2 v2.0.0-20211213093848-f0fd3099c235
	chainmaker.org/chainmaker/consensus-solo/v2 v2.0.0-20211210081341-96bcb7aa4b9c
	chainmaker.org/chainmaker/consensus-tbft/v2 v2.0.0-20211213072659-c9d9a2ad2726
	chainmaker.org/chainmaker/consensus-utils/v2 v2.0.0-20211210075517-b641978ff2fc
	chainmaker.org/chainmaker/localconf/v2 v2.1.1-0.20211110030026-ce2a7f3760cd
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/net-common v0.0.7-0.20211109085844-739f0f904b96
	chainmaker.org/chainmaker/net-libp2p v1.0.1-0.20211109090515-4889a63c74af
	chainmaker.org/chainmaker/net-liquid v1.0.1-0.20211122114338-22ed0765724f
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211213065726-1e0a814347b6
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211207133449-dcd91d474c06
	chainmaker.org/chainmaker/sdk-go/v2 v2.0.1-0.20211213072132-b1425a5dc764
	chainmaker.org/chainmaker/store-sqldb/v2 v2.1.1-0.20211202031309-159eb01a87b1 // indirect
	chainmaker.org/chainmaker/store/v2 v2.1.1-0.20211213085525-c037f5fb5f6d
	chainmaker.org/chainmaker/txpool-batch/v2 v2.1.1-0.20211129022941-e7a476018d0c
	chainmaker.org/chainmaker/txpool-single/v2 v2.1.1-0.20211109075506-aea78872cdc6
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20211208040318-ad58981b7d09
	chainmaker.org/chainmaker/vm-docker-go v0.0.0-20211207085346-9fd71eabcc42
	chainmaker.org/chainmaker/vm-evm/v2 v2.0.0-20211210095634-f8bc080855bf
	chainmaker.org/chainmaker/vm-gasm/v2 v2.1.1-0.20211210095537-dd17ca44b22f
	chainmaker.org/chainmaker/vm-native/v2 v2.1.1-0.20211213072243-6cb0732b0787
	chainmaker.org/chainmaker/vm-wasmer/v2 v2.1.1-0.20211210095550-20541baca032
	chainmaker.org/chainmaker/vm-wxvm/v2 v2.1.1-0.20211214062844-e41c8f17122f // indirect
	chainmaker.org/chainmaker/vm/v2 v2.1.1-0.20211210095054-3cb65a1ecfc8
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
	github.com/mitchellh/mapstructure v1.4.2
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
	google.golang.org/genproto v0.0.0-20210828152312-66f60bf46e71 // indirect
	google.golang.org/grpc v1.41.0
	gorm.io/driver/mysql v1.2.0
	gorm.io/gorm v1.22.3
)

replace (
	github.com/libp2p/go-libp2p-core => chainmaker.org/chainmaker/libp2p-core v0.0.2
	github.com/spf13/afero => github.com/spf13/afero v1.5.1 //for go1.15 build
	github.com/spf13/viper => github.com/spf13/viper v1.7.1 //for go1.15 build
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)
