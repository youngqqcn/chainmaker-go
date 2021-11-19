module chainmaker.org/chainmaker-go/blockchain

go 1.15

require (
	chainmaker.org/chainmaker-go/accesscontrol v0.0.0
	chainmaker.org/chainmaker-go/consensus v0.0.0
	chainmaker.org/chainmaker-go/core v0.0.0-00010101000000-000000000000
	chainmaker.org/chainmaker-go/net v0.0.0
	chainmaker.org/chainmaker-go/snapshot v0.0.0
	chainmaker.org/chainmaker-go/subscriber v0.0.0
	chainmaker.org/chainmaker-go/sync v0.0.0
	chainmaker.org/chainmaker-go/txpool v0.0.0
	chainmaker.org/chainmaker-go/vm v0.0.0
	chainmaker.org/chainmaker/chainconf/v2 v2.1.1-0.20211109075405-cc95de741f5e
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211117131805-630800bfd361
	chainmaker.org/chainmaker/consensus-utils/v2 v2.0.0-20211115084213-42e840e1efee
	chainmaker.org/chainmaker/localconf/v2 v2.1.1-0.20211110030026-ce2a7f3760cd
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211117140137-9095a7ab7a69
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211119081550-ff85fecc318d
	chainmaker.org/chainmaker/store/v2 v2.1.1-0.20211118114249-03894d6161cb
	chainmaker.org/chainmaker/txpool-batch/v2 v2.1.1-0.20211109075600-a0a811fe0650 // indirect
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20211117144316-3f4a940e94f0
	chainmaker.org/chainmaker/vm v0.0.0-20211028094551-9b2cca96d10d
	github.com/fatih/color v1.13.0 // indirect
	github.com/hokaccha/go-prettyjson v0.0.0-20210113012101-fb4e108d2519 // indirect
	github.com/mattn/go-colorable v0.1.11 // indirect
	github.com/mitchellh/mapstructure v1.4.2
)

replace (
	chainmaker.org/chainmaker-go/accesscontrol => ../accesscontrol
	chainmaker.org/chainmaker-go/consensus => ../consensus
	chainmaker.org/chainmaker-go/core => ../core
	chainmaker.org/chainmaker-go/net => ../net
	chainmaker.org/chainmaker-go/snapshot => ../snapshot
	chainmaker.org/chainmaker-go/subscriber => ../subscriber
	chainmaker.org/chainmaker-go/sync => ../sync
	chainmaker.org/chainmaker-go/txpool => ../txpool
	chainmaker.org/chainmaker-go/vm => ../vm
	github.com/libp2p/go-libp2p-core => chainmaker.org/chainmaker/libp2p-core v0.0.2
	github.com/spf13/viper => github.com/spf13/viper v1.7.1 //for go1.15 build
	google.golang.org/grpc v1.40.0 => google.golang.org/grpc v1.26.0
)
