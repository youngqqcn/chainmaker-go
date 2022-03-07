module chainmaker.org/chainmaker-go

go 1.15

require (
	chainmaker.org/chainmaker/chainconf/v2 v2.1.2-0.20220304064458-4d8be01d6961
	chainmaker.org/chainmaker/common/v2 v2.1.2-0.20220301062133-c3850a6efa7f
	chainmaker.org/chainmaker/consensus-dpos/v2 v2.0.0-20211220115802-ad63e1565a38
	chainmaker.org/chainmaker/consensus-maxbft/v2 v2.0.0-20220214094627-a616ca03d6aa
	chainmaker.org/chainmaker/consensus-raft/v2 v2.0.0-20220221033200-bcc78c45c616
	chainmaker.org/chainmaker/consensus-solo/v2 v2.0.0-20211220120848-36f05d90fd9c
	chainmaker.org/chainmaker/consensus-tbft/v2 v2.0.0-20220125101157-16e2bb0098ae
	chainmaker.org/chainmaker/consensus-utils/v2 v2.0.0-20220215071220-a392a07ce38a
	chainmaker.org/chainmaker/localconf/v2 v2.1.1-0.20220302084404-928bd4a6131a
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20220128022235-c984177a37cc
	chainmaker.org/chainmaker/net-common v1.0.2-0.20220118021419-d8b87c4d6f72
	chainmaker.org/chainmaker/net-libp2p v1.0.2-0.20220119112630-026b97cb6878
	chainmaker.org/chainmaker/net-liquid v1.0.1
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20220304062644-4fc635a393ba
	chainmaker.org/chainmaker/protocol/v2 v2.1.2-0.20220304060945-5ff4c843652c
	chainmaker.org/chainmaker/sdk-go/v2 v2.0.1-0.20220307094104-7c57adca66c3
	chainmaker.org/chainmaker/store/v2 v2.1.1-0.20220228032811-8168a2644c25
	chainmaker.org/chainmaker/txpool-batch/v2 v2.1.1-0.20220114030910-a28969c65faf
	chainmaker.org/chainmaker/txpool-single/v2 v2.1.1-0.20220228060609-44f14d1e255e
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20220124072923-e59e8500c1a7
	chainmaker.org/chainmaker/vm-docker-go/v2 v2.1.1-0.20220304113245-ebc00b54f82d
	chainmaker.org/chainmaker/vm-evm/v2 v2.1.1-0.20220124125450-af7817cb999a
	chainmaker.org/chainmaker/vm-gasm/v2 v2.1.1-0.20220223062538-5503a7415fe3
	chainmaker.org/chainmaker/vm-native/v2 v2.1.2-0.20220304082303-f8f862054bad
	chainmaker.org/chainmaker/vm-wasmer/v2 v2.1.1-0.20220224064011-5a7caccf53ed
	chainmaker.org/chainmaker/vm-wxvm/v2 v2.1.1-0.20211223061926-78b8d34d3aa3
	chainmaker.org/chainmaker/vm/v2 v2.1.2-0.20220304070757-170dbcfb5bd5
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
