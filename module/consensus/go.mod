module chainmaker.org/chainmaker-go/consensus

go 1.15

require (
	chainmaker.org/chainmaker-go/accesscontrol v0.0.0-00010101000000-000000000000
	chainmaker.org/chainmaker/chainconf/v2 v2.1.1-0.20211109075405-cc95de741f5e
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211117131805-630800bfd361
	chainmaker.org/chainmaker/localconf/v2 v2.1.1-0.20211110030026-ce2a7f3760cd
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211110082822-61bdedd084bd
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211116092258-b0de845d438c
	chainmaker.org/chainmaker/raftwal/v2 v2.0.3
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20211109074701-81d58330e787
	chainmaker.org/chainmaker/vm-native v0.0.0-20211029075817-046b5f1bbc31
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
	github.com/libp2p/go-libp2p-core => chainmaker.org/chainmaker/libp2p-core v0.0.2
	github.com/spf13/viper => github.com/spf13/viper v1.7.1 //for go1.15 build
)
