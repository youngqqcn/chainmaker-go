module chainmaker.org/chainmaker-go/core

go 1.15

require (
	chainmaker.org/chainmaker-go/subscriber v0.0.0
	chainmaker.org/chainmaker/chainconf/v2 v2.1.1-0.20211109075405-cc95de741f5e
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211117131805-630800bfd361
	chainmaker.org/chainmaker/consensus/v2 v2.0.0-20211119102054-ff9002f55149
	chainmaker.org/chainmaker/localconf/v2 v2.1.1-0.20211110030026-ce2a7f3760cd
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211117140137-9095a7ab7a69
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211119081550-ff85fecc318d
	chainmaker.org/chainmaker/txpool-batch/v2 v2.1.1-0.20211109075600-a0a811fe0650
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20211117144316-3f4a940e94f0
	chainmaker.org/chainmaker/vm/v2 v2.1.1-0.20211117062641-a1b17375caf8
	github.com/ethereum/go-ethereum v1.10.3 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mitchellh/mapstructure v1.4.2 // indirect
	github.com/panjf2000/ants/v2 v2.4.3
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/stretchr/testify v1.7.0
	gopkg.in/ini.v1 v1.63.2 // indirect
)

replace (
	chainmaker.org/chainmaker-go/accesscontrol => ../accesscontrol
	chainmaker.org/chainmaker-go/consensus => ../consensus
	chainmaker.org/chainmaker-go/consensus/dpos => ./../consensus/dpos

	chainmaker.org/chainmaker-go/monitor => ../monitor
	chainmaker.org/chainmaker-go/subscriber => ../subscriber
	github.com/spf13/viper => github.com/spf13/viper v1.7.1 //for go1.15 build
)
