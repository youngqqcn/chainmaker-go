module chainmaker.org/chainmaker-go/consensus

go 1.15

require (
	chainmaker.org/chainmaker-go/accesscontrol v0.0.0
	chainmaker.org/chainmaker/chainconf/v2 v2.1.1-0.20211110023535-5bf814f90c4c
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211125080705-8c4f85e6ca19
	chainmaker.org/chainmaker/localconf/v2 v2.1.0
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211125125051-7aa6e0da2379
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211125064016-15cf21479e69
	chainmaker.org/chainmaker/raftwal/v2 v2.1.0
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20211109074701-81d58330e787
	chainmaker.org/chainmaker/vm-native/v2 v2.1.1-0.20211125125920-75b942f495cc
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/golang/mock v1.6.0
	github.com/pingcap/errors v0.11.5-0.20201126102027-b0a155152ca3 // indirect
	github.com/pingcap/log v0.0.0-20201112100606-8f1e84a3abc8 // indirect
	github.com/shirou/gopsutil v3.21.4-0.20210419000835-c7a38de76ee5+incompatible // indirect
	github.com/spf13/viper v1.9.0
	github.com/stretchr/testify v1.7.0
	github.com/studyzy/sqlparse v0.0.0-20210525032257-e7b9574609c3 // indirect
	github.com/syndtr/goleveldb v1.0.1-0.20210305035536-64b5b1c73954
	github.com/thoas/go-funk v0.8.0
	go.etcd.io/etcd/client/pkg/v3 v3.5.0
	go.etcd.io/etcd/raft/v3 v3.5.0
	go.etcd.io/etcd/server/v3 v3.5.0
	go.uber.org/zap v1.19.1
)

replace (
	chainmaker.org/chainmaker-go/accesscontrol => ../accesscontrol
	github.com/libp2p/go-libp2p-core => chainmaker.org/chainmaker/libp2p-core v1.0.0
)
