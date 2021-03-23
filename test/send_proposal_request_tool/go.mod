module chainmaker.org/chainmaker-go/test/send_proposal_request_tool

go 1.15

require (
	chainmaker.org/chainmaker-go/accesscontrol v0.0.0
	chainmaker.org/chainmaker-go/common v0.0.0
	chainmaker.org/chainmaker-go/logger v0.0.0
	chainmaker.org/chainmaker-go/pb/protogo v0.0.0
	chainmaker.org/chainmaker-go/protocol v0.0.0
	chainmaker.org/chainmaker-go/utils v0.0.0
	github.com/gogo/protobuf v1.3.2
	github.com/spf13/cobra v1.1.1
	google.golang.org/grpc v1.36.0
)

replace (
	chainmaker.org/chainmaker-go/accesscontrol => ../../module/accesscontrol
	chainmaker.org/chainmaker-go/common => ../../common
	chainmaker.org/chainmaker-go/localconf => ../../module/conf/localconf
	chainmaker.org/chainmaker-go/logger => ../../module/logger
	chainmaker.org/chainmaker-go/pb/protogo => ../../pb/protogo
	chainmaker.org/chainmaker-go/protocol => ../../protocol
	chainmaker.org/chainmaker-go/utils => ../../module/utils
)