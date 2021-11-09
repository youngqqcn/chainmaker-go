module chainmaker.org/chainmaker-go/blockchain

go 1.15

require (
	chainmaker.org/chainmaker-go/accesscontrol v0.0.0
	chainmaker.org/chainmaker-go/net v0.0.0
	chainmaker.org/chainmaker-go/snapshot v0.0.0
	chainmaker.org/chainmaker-go/subscriber v0.0.0
	chainmaker.org/chainmaker-go/sync v0.0.0
	chainmaker.org/chainmaker-go/txpool v0.0.0
	chainmaker.org/chainmaker-go/vm v0.0.0
	chainmaker.org/chainmaker/chainconf/v2 v2.0.0-20211108091950-b471f3c7ce3d
	chainmaker.org/chainmaker/common/v2 v2.1.0
	chainmaker.org/chainmaker/localconf/v2 v2.0.0-20211101111610-0d268248b5c8
	chainmaker.org/chainmaker/logger/v2 v2.1.0
	chainmaker.org/chainmaker/pb-go/v2 v2.1.0
	chainmaker.org/chainmaker/protocol/v2 v2.0.1-0.20211108075639-576c31f03396
	chainmaker.org/chainmaker/store/v2 v2.0.1-0.20211101060552-ab2318ad9c6e
	chainmaker.org/chainmaker/txpool-batch/v2 v2.0.0-20211019074609-46e3d29f0908 // indirect
	chainmaker.org/chainmaker/utils/v2 v2.0.0-20211108092352-2a3335a4ba15
	chainmaker.org/chainmaker/vm v0.0.0-20211028094551-9b2cca96d10d
	github.com/bketelsen/crypt v0.0.4 // indirect
	github.com/ethereum/go-ethereum v1.10.3 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20210113012101-fb4e108d2519 // indirect
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-colorable v0.1.11 // indirect
	github.com/mitchellh/mapstructure v1.4.2
	github.com/panjf2000/ants/v2 v2.4.3 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/sagikazarmark/crypt v0.1.0 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/cobra v1.1.3 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1 // indirect
	gopkg.in/ini.v1 v1.63.2 // indirect
)

replace (
	chainmaker.org/chainmaker-go/accesscontrol => ../accesscontrol
	chainmaker.org/chainmaker-go/core => ../core
	chainmaker.org/chainmaker-go/net => ../net
	chainmaker.org/chainmaker-go/snapshot => ../snapshot
	chainmaker.org/chainmaker-go/subscriber => ../subscriber
	chainmaker.org/chainmaker-go/sync => ../sync
	chainmaker.org/chainmaker-go/txpool => ../txpool
	chainmaker.org/chainmaker-go/vm => ../vm
	github.com/libp2p/go-libp2p-core => chainmaker.org/chainmaker/libp2p-core v0.0.2
	google.golang.org/grpc v1.40.0 => google.golang.org/grpc v1.26.0
)
