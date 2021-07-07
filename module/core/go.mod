module chainmaker.org/chainmaker-go/core

go 1.15

require (
	chainmaker.org/chainmaker-go/chainconf v0.0.0
	chainmaker.org/chainmaker-go/consensus v0.0.0
	chainmaker.org/chainmaker-go/localconf v0.0.0
	chainmaker.org/chainmaker-go/logger v0.0.0
	chainmaker.org/chainmaker-go/monitor v0.0.0
	chainmaker.org/chainmaker-go/store v0.0.0
	chainmaker.org/chainmaker-go/subscriber v0.0.0
	chainmaker.org/chainmaker-go/utils v0.0.0
	chainmaker.org/chainmaker/common v0.0.0-20210630062216-42b826d5ecea
	chainmaker.org/chainmaker/pb-go v0.0.0-20210630123814-27dfc86893cf
	chainmaker.org/chainmaker/protocol v0.0.0-20210630150922-63cd365c6511
	github.com/gogo/protobuf v1.3.2
	github.com/google/martian v2.1.0+incompatible
	github.com/panjf2000/ants/v2 v2.4.3
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/common v0.15.0
	github.com/stretchr/testify v1.7.0
)

replace (
	chainmaker.org/chainmaker-go/accesscontrol => ../accesscontrol
	chainmaker.org/chainmaker-go/chainconf => ./../conf/chainconf

	chainmaker.org/chainmaker-go/consensus => ../consensus
	chainmaker.org/chainmaker-go/dpos => ../dpos
	chainmaker.org/chainmaker-go/evm => ./../../module/vm/evm
	chainmaker.org/chainmaker-go/gasm => ./../../module/vm/gasm
	chainmaker.org/chainmaker-go/localconf => ./../conf/localconf
	chainmaker.org/chainmaker-go/logger => ../logger

	chainmaker.org/chainmaker-go/monitor => ../monitor

	chainmaker.org/chainmaker-go/store => ../store
	chainmaker.org/chainmaker-go/subscriber => ../subscriber
	chainmaker.org/chainmaker-go/utils => ../utils
	chainmaker.org/chainmaker-go/vm => ./../../module/vm
	chainmaker.org/chainmaker-go/wasi => ./../../module/vm/wasi
	chainmaker.org/chainmaker-go/wasmer => ./../../module/vm/wasmer
	chainmaker.org/chainmaker-go/wxvm => ./../../module/vm/wxvm
)
