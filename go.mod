module chainmaker.org/chainmaker-go

go 1.15

require (
	chainmaker.org/chainmaker/chainconf/v2 v2.1.1-0.20211129064036-3ae6487c7770
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211117131805-630800bfd361
	chainmaker.org/chainmaker/consensus-chainedbft/v2 v2.0.0-20211123115807-5d8e35647df4
	chainmaker.org/chainmaker/consensus-dpos/v2 v2.0.0-20211119100858-b806ddbb6d35
	chainmaker.org/chainmaker/consensus-raft/v2 v2.0.0-20211119100947-ce17472c410c
	chainmaker.org/chainmaker/consensus-solo/v2 v2.0.0-20211123123020-d0f19f91f625
	chainmaker.org/chainmaker/consensus-tbft/v2 v2.0.0-20211119101400-96829b485267
	chainmaker.org/chainmaker/consensus-utils/v2 v2.0.0-20211115084213-42e840e1efee
	chainmaker.org/chainmaker/consensus/v2 v2.0.0-20211123121814-dd9c430b5a7b
	chainmaker.org/chainmaker/localconf/v2 v2.1.1-0.20211110030026-ce2a7f3760cd
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/net-common v0.0.7-0.20211109085844-739f0f904b96
	chainmaker.org/chainmaker/net-libp2p v1.0.1-0.20211109090515-4889a63c74af
	chainmaker.org/chainmaker/net-liquid v1.0.1-0.20211122114338-22ed0765724f
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211117140137-9095a7ab7a69
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211119081550-ff85fecc318d
	chainmaker.org/chainmaker/sdk-go/v2 v2.0.1-0.20211122084724-d9ebcd942459
	chainmaker.org/chainmaker/store/v2 v2.1.1-0.20211202031657-09bc7ab7fac5
	chainmaker.org/chainmaker/txpool-batch/v2 v2.1.1-0.20211109075600-a0a811fe0650
	chainmaker.org/chainmaker/txpool-single/v2 v2.1.1-0.20211109075506-aea78872cdc6
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20211117144316-3f4a940e94f0
	chainmaker.org/chainmaker/vm-evm/v2 v2.1.1-0.20211109084614-e0e363b49d47
	chainmaker.org/chainmaker/vm-gasm/v2 v2.1.1-0.20211109083312-3f36d5d4e4d2
	chainmaker.org/chainmaker/vm-native/v2 v2.1.1-0.20211203151445-3e9efc3f8ad6
	chainmaker.org/chainmaker/vm-wasmer/v2 v2.1.1-0.20211119040808-92b65899613e
	chainmaker.org/chainmaker/vm-wxvm/v2 v2.1.1-0.20211115073737-27720f7ad7d1
	chainmaker.org/chainmaker/vm/v2 v2.1.1-0.20211203152046-4bda897ae501
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
	golang.org/x/crypto v0.0.0-20210817164053-32db794688a5
	golang.org/x/time v0.0.0-20210220033141-f8bda1e9f3ba
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
