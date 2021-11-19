module chainmaker.org/chainmaker-go/consensus

go 1.15

require (
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211117131805-630800bfd361
	chainmaker.org/chainmaker/consensus-dpos/v2 v2.0.0-20211119100858-b806ddbb6d35
	chainmaker.org/chainmaker/consensus-hotstuff/v2 v2.0.0-20211119101639-ae06999f7408
	chainmaker.org/chainmaker/consensus-raft/v2 v2.0.0-20211119100947-ce17472c410c
	chainmaker.org/chainmaker/consensus-solo/v2 v2.0.0-20211119095958-4b1cce193b32
	chainmaker.org/chainmaker/consensus-tbft/v2 v2.0.0-20211119101400-96829b485267
	chainmaker.org/chainmaker/consensus-utils/v2 v2.0.0-20211115084213-42e840e1efee
	chainmaker.org/chainmaker/localconf/v2 v2.1.1-0.20211110030026-ce2a7f3760cd
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211117140137-9095a7ab7a69
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211119081550-ff85fecc318d
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/mock v1.6.0
	github.com/magiconair/properties v1.8.5 // indirect
	github.com/mattn/go-colorable v0.1.6 // indirect
	github.com/mitchellh/mapstructure v1.4.2 // indirect
	github.com/pelletier/go-toml v1.9.4 // indirect
	github.com/spf13/afero v1.6.0 // indirect
	github.com/spf13/cast v1.4.1 // indirect
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	gopkg.in/ini.v1 v1.63.2 // indirect
)

replace github.com/spf13/viper => github.com/spf13/viper v1.7.1 //for go1.15 build
