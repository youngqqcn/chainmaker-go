module chainmaker.org/chainmaker-go

go 1.15

require (
	chainmaker.org/chainmaker/chainconf/v2 v2.1.2-0.20220105094659-53b31f1ad2bc
	chainmaker.org/chainmaker/common/v2 v2.1.2-0.20220102143903-cc96a7c51ff4
	chainmaker.org/chainmaker/consensus-chainedbft/v2 v2.0.0-20211214032204-5f155aeeb5a2
	chainmaker.org/chainmaker/consensus-dpos/v2 v2.0.0-20211220115802-ad63e1565a38
	chainmaker.org/chainmaker/consensus-maxbft/v2 v2.0.0-20211220120509-ffe11f1c1a25
	chainmaker.org/chainmaker/consensus-raft/v2 v2.0.0-20211231065915-33ec8814372a
	chainmaker.org/chainmaker/consensus-solo/v2 v2.0.0-20211220120848-36f05d90fd9c
	chainmaker.org/chainmaker/consensus-tbft/v2 v2.0.0-20220107054455-20db25508ea6
	chainmaker.org/chainmaker/consensus-utils/v2 v2.0.0-20220106130814-c0ad4f98d881
	chainmaker.org/chainmaker/localconf/v2 v2.1.1-0.20211230035526-c3c28e290ca4
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211214124250-621f11b35ab0
	chainmaker.org/chainmaker/net-common v1.0.1
	chainmaker.org/chainmaker/net-libp2p v1.0.2-0.20220107085327-8fbe1cd0d391
	chainmaker.org/chainmaker/net-liquid v1.0.1-0.20211214135953-619e26ee9681
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211230035615-3d074e0887c6
	chainmaker.org/chainmaker/protocol/v2 v2.1.2-0.20211230130206-933dddf33257
	chainmaker.org/chainmaker/sdk-go/v2 v2.0.1-0.20211215093913-d1edc3f9299e
	chainmaker.org/chainmaker/store/v2 v2.1.1-0.20220107080206-397a0e0ff2aa
	chainmaker.org/chainmaker/txpool-batch/v2 v2.1.1-0.20220107091903-40e9e6dc62ad
	chainmaker.org/chainmaker/txpool-single/v2 v2.1.1-0.20211214125452-9037b3ed764a
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20220107071133-037a6a955af5
	chainmaker.org/chainmaker/vm-docker-go/v2 v2.1.1-0.20211230112054-20b916b596f5
	chainmaker.org/chainmaker/vm-evm/v2 v2.1.1-0.20211222085520-073bb7401465
	chainmaker.org/chainmaker/vm-gasm/v2 v2.1.1-0.20211221065927-2b2979cbe7ae
	chainmaker.org/chainmaker/vm-native/v2 v2.1.2-0.20220106080644-19ea27fd7007
	chainmaker.org/chainmaker/vm-wasmer/v2 v2.1.1-0.20211228150309-955bc04b6557
	chainmaker.org/chainmaker/vm-wxvm/v2 v2.1.1-0.20211223061926-78b8d34d3aa3
	chainmaker.org/chainmaker/vm/v2 v2.1.1-0.20211215035257-059dbafc381f
	code.cloudfoundry.org/bytefmt v0.0.0-20211005130812-5bb3c17173e5
	github.com/Rican7/retry v0.1.0
	github.com/Workiva/go-datastructures v1.0.53
	github.com/c-bata/go-prompt v0.2.2
	github.com/common-nighthawk/go-figure v0.0.0-20210622060536-734e95fb86be
	github.com/ethereum/go-ethereum v1.10.4
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/google/go-cmp v0.5.6 // indirect
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
	github.com/spf13/viper v1.10.1
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
