module chainmaker.org/chainmaker-go/test/send_proposal_request_tool

go 1.15

require (
	chainmaker.org/chainmaker-go/accesscontrol v0.0.0
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211117131805-630800bfd361
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211110082822-61bdedd084bd
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211116092258-b0de845d438c
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20211109074701-81d58330e787
	github.com/Rican7/retry v0.1.0
	github.com/ethereum/go-ethereum v1.10.2
	github.com/gogo/protobuf v1.3.2
	github.com/mr-tron/base58 v1.2.0
	github.com/spf13/cobra v1.1.1
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/pretty v1.2.0
	google.golang.org/grpc v1.40.0
)

replace (
	chainmaker.org/chainmaker-go/accesscontrol => ../../module/accesscontrol

	chainmaker.org/chainmaker-go/localconf => ../../module/conf/localconf
	chainmaker.org/chainmaker-go/logger => ../../module/logger

)
