module chainmaker.org/chainmaker-go/wasmer

go 1.15

require (
	chainmaker.org/chainmaker-go/logger v0.0.0
	chainmaker.org/chainmaker-go/utils v0.0.0
	chainmaker.org/chainmaker-go/wasi v0.0.0
	chainmaker.org/chainmaker/common v0.0.0-20210811075857-d3b57d983071
	chainmaker.org/chainmaker/pb-go v0.0.0-20210812071728-a0312d5ac16a
	chainmaker.org/chainmaker/protocol v0.0.0-20210810081254-4947fb9a5306
)

replace (
	chainmaker.org/chainmaker-go/localconf => ../../conf/localconf
	chainmaker.org/chainmaker-go/logger => ../../logger
	chainmaker.org/chainmaker-go/store => ../../store

	chainmaker.org/chainmaker-go/utils => ../../utils
	chainmaker.org/chainmaker-go/wasi => ../wasi
)
