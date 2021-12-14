module chainmaker.org/chainmaker-go

go 1.15

require (
	chainmaker.org/chainmaker/chainconf/v2 v2.1.1-0.20211214114856-83c398b3db30
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211214041159-fe0b2240f08c
	chainmaker.org/chainmaker/consensus-chainedbft/v2 v2.0.0-20211207134138-c082d96d81e8
	chainmaker.org/chainmaker/consensus-dpos/v2 v2.0.0-20211214033908-a9183a7b963a
	chainmaker.org/chainmaker/consensus-maxbft/v2 v2.0.0-20211210081411-5ef693cd806f
	chainmaker.org/chainmaker/consensus-raft/v2 v2.0.0-20211213100300-17d8e769f98e
	chainmaker.org/chainmaker/consensus-solo/v2 v2.0.0-20211214030920-ea0e463ad738
	chainmaker.org/chainmaker/consensus-tbft/v2 v2.0.0-20211214120032-9eb483ea4ef5
	chainmaker.org/chainmaker/consensus-utils/v2 v2.0.0-20211214095113-179cc53746eb
	chainmaker.org/chainmaker/localconf/v2 v2.1.1-0.20211213123219-8e2e0cdcd628
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/net-common v0.0.7-0.20211109085844-739f0f904b96
	chainmaker.org/chainmaker/net-libp2p v1.0.1-0.20211213064428-ff6ab75341e1
	chainmaker.org/chainmaker/net-liquid v1.0.1-0.20211122114338-22ed0765724f
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211214073209-1288f39035e8
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211214120754-62a8fab32ce9
	chainmaker.org/chainmaker/sdk-go/v2 v2.0.1-0.20211214070956-cbd556d24c45
	chainmaker.org/chainmaker/store/v2 v2.1.1-0.20211214064027-a3882620f281
	chainmaker.org/chainmaker/txpool-batch/v2 v2.1.1-0.20211129022941-e7a476018d0c
	chainmaker.org/chainmaker/txpool-single/v2 v2.1.1-0.20211109075506-aea78872cdc6
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20211214041912-2ba2b0c9b83d
	chainmaker.org/chainmaker/vm-docker-go v0.0.0-20211214063103-6f6b10d166d2
	chainmaker.org/chainmaker/vm-evm/v2 v2.1.1-0.20211214063845-92b7c135504e
	chainmaker.org/chainmaker/vm-gasm/v2 v2.1.1-0.20211214075335-8b840f5c2760
	chainmaker.org/chainmaker/vm-native/v2 v2.1.1-0.20211214071729-b76749c54788
	chainmaker.org/chainmaker/vm-wasmer/v2 v2.1.1-0.20211214095207-0071a6c50211
	chainmaker.org/chainmaker/vm-wxvm/v2 v2.1.1-0.20211214062844-e41c8f17122f
	chainmaker.org/chainmaker/vm/v2 v2.1.1-0.20211214072548-7aa5b21b4923
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
