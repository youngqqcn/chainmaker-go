module chainmaker.org/chainmaker-go

go 1.15

require (
	chainmaker.org/chainmaker/chainconf/v2 v2.1.2-0.20220113072252-aecb4f7ffef0
	chainmaker.org/chainmaker/common/v2 v2.1.2-0.20220118101118-f256fc9c55e6
	chainmaker.org/chainmaker/consensus-dpos/v2 v2.0.0-20211220115802-ad63e1565a38
	chainmaker.org/chainmaker/consensus-maxbft/v2 v2.0.0-20220111123612-a8943e389db5
	chainmaker.org/chainmaker/consensus-raft/v2 v2.0.0-20211231065915-33ec8814372a
	chainmaker.org/chainmaker/consensus-solo/v2 v2.0.0-20211220120848-36f05d90fd9c
	chainmaker.org/chainmaker/consensus-tbft/v2 v2.0.0-20220111100749-02365fea8179
	chainmaker.org/chainmaker/consensus-utils/v2 v2.0.0-20220111072256-e5816e4f8aae
	chainmaker.org/chainmaker/localconf/v2 v2.1.1-0.20220112085516-908b8478be8f
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211214124250-621f11b35ab0
	chainmaker.org/chainmaker/net-common v1.0.2-0.20220118021419-d8b87c4d6f72
	chainmaker.org/chainmaker/net-libp2p v1.0.2-0.20220119112630-026b97cb6878
	chainmaker.org/chainmaker/net-liquid v1.0.1
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20220119062501-d1b3a201f7fb
	chainmaker.org/chainmaker/protocol/v2 v2.1.2-0.20220113081648-bbf6c1946b59
	chainmaker.org/chainmaker/sdk-go/v2 v2.0.1-0.20220118082926-a5acace65c33
	chainmaker.org/chainmaker/store/v2 v2.1.1-0.20220114135457-ce8272da8e67
	chainmaker.org/chainmaker/txpool-batch/v2 v2.1.1-0.20220114030910-a28969c65faf
	chainmaker.org/chainmaker/txpool-single/v2 v2.1.1-0.20220113122238-1a474b15dd18
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20220114120415-8e1af1e262a7
	chainmaker.org/chainmaker/vm-docker-go/v2 v2.1.1-0.20220111073631-56233e995445
	chainmaker.org/chainmaker/vm-evm/v2 v2.1.1-0.20211222085520-073bb7401465
	chainmaker.org/chainmaker/vm-gasm/v2 v2.1.1-0.20220118034122-43ec8bd22ed0
	chainmaker.org/chainmaker/vm-native/v2 v2.1.2-0.20220120055651-0140045e308b
	chainmaker.org/chainmaker/vm-wasmer/v2 v2.1.1-0.20211228150309-955bc04b6557
	chainmaker.org/chainmaker/vm-wxvm/v2 v2.1.1-0.20211223061926-78b8d34d3aa3
	chainmaker.org/chainmaker/vm/v2 v2.1.2-0.20220120060006-45ee42c292e0
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
	github.com/panjf2000/ants/v2 v2.4.7
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
