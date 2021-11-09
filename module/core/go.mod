module chainmaker.org/chainmaker-go/core

go 1.15

require (
	chainmaker.org/chainmaker-go/subscriber v0.0.0
	chainmaker.org/chainmaker/chainconf/v2 v2.0.0-20211108091950-b471f3c7ce3d
	chainmaker.org/chainmaker/common/v2 v2.1.0
	chainmaker.org/chainmaker/consensus-dpos/v2 v2.0.0-20211112092735-d15ea84c5f44 // indirect
	chainmaker.org/chainmaker/consensus-raft/v2 v2.0.0-20211112163240-bd1b63cc16bb // indirect
	chainmaker.org/chainmaker/consensus-solo/v2 v2.0.0-20211112091638-d0d658ddbdfa // indirect
	chainmaker.org/chainmaker/consensus/v2 v2.0.0-20211112163732-66772d402b44 // indirect
	chainmaker.org/chainmaker/localconf/v2 v2.0.0-20211101111610-0d268248b5c8
	chainmaker.org/chainmaker/logger/v2 v2.1.0
	chainmaker.org/chainmaker/pb-go/v2 v2.1.0
	chainmaker.org/chainmaker/protocol/v2 v2.0.1-0.20211108075639-576c31f03396
	chainmaker.org/chainmaker/raftwal/v2 v2.0.3 // indirect
	chainmaker.org/chainmaker/txpool-batch/v2 v2.0.0-20211019074609-46e3d29f0908
	chainmaker.org/chainmaker/utils/v2 v2.0.0-20211108092352-2a3335a4ba15
	chainmaker.org/chainmaker/vm v0.0.0-20211028094551-9b2cca96d10d
	github.com/ethereum/go-ethereum v1.10.3 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mitchellh/mapstructure v1.4.2 // indirect
	github.com/panjf2000/ants/v2 v2.4.3
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/sagikazarmark/crypt v0.1.0 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/thoas/go-funk v0.9.1 // indirect
	go.etcd.io/etcd/server/v3 v3.5.1 // indirect
	google.golang.org/grpc v1.40.0 // indirect
	gopkg.in/ini.v1 v1.63.2 // indirect
)

replace chainmaker.org/chainmaker-go/subscriber => ../subscriber
