module chainmaker.org/chainmaker-go/core

go 1.15

require (
	chainmaker.org/chainmaker-go/consensus v0.0.0
	chainmaker.org/chainmaker-go/subscriber v0.0.0
	chainmaker.org/chainmaker/chainconf/v2 v2.1.1-0.20211110023535-5bf814f90c4c
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211108032042-88ed7467a15c
	chainmaker.org/chainmaker/localconf/v2 v2.1.0
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211122052606-1ea3f17343e7
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211117024857-2641037a7269
	chainmaker.org/chainmaker/txpool-batch/v2 v2.1.0
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20211109074701-81d58330e787
	chainmaker.org/chainmaker/vm-native/v2 v2.1.1-0.20211122122336-75d168875de2
	chainmaker.org/chainmaker/vm/v2 v2.1.0
	github.com/ethereum/go-ethereum v1.10.3 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/panjf2000/ants/v2 v2.4.6
	github.com/prometheus/client_golang v1.11.0
	github.com/stretchr/testify v1.7.0
	google.golang.org/grpc v1.40.0 // indirect
)

replace (
	chainmaker.org/chainmaker-go/accesscontrol => ../accesscontrol
	chainmaker.org/chainmaker-go/consensus => ../consensus
	chainmaker.org/chainmaker-go/consensus/dpos => ./../consensus/dpos
	chainmaker.org/chainmaker-go/subscriber => ../subscriber
)
