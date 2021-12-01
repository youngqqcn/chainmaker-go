module chainmaker.org/chainmaker-go/core

go 1.15

require (
	chainmaker.org/chainmaker-go/consensus v0.0.0
	chainmaker.org/chainmaker-go/subscriber v0.0.0-00010101000000-000000000000
	chainmaker.org/chainmaker/chainconf/v2 v2.1.1-0.20211201032240-900b959567cf
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211125080705-8c4f85e6ca19
	chainmaker.org/chainmaker/localconf/v2 v2.1.0
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211130072802-e9d8fae57083
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211130023410-5df7eb63bfb8
	chainmaker.org/chainmaker/txpool-batch/v2 v2.1.0
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20211130112814-b3126608050b
	chainmaker.org/chainmaker/vm-native/v2 v2.1.1-0.20211201072525-3484cf854cde
	chainmaker.org/chainmaker/vm/v2 v2.1.1-0.20211201074641-57d2383c0b29
	github.com/gogo/protobuf v1.3.2
	github.com/holiman/uint256 v1.2.0
	github.com/panjf2000/ants/v2 v2.4.6
	github.com/prometheus/client_golang v1.11.0
	github.com/stretchr/testify v1.7.0
)

replace (
	chainmaker.org/chainmaker-go/accesscontrol => ../accesscontrol
	chainmaker.org/chainmaker-go/consensus => ../consensus
	chainmaker.org/chainmaker-go/consensus/dpos => ./../consensus/dpos
	chainmaker.org/chainmaker-go/subscriber => ../subscriber
)
