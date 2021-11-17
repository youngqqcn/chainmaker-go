module chainmaker.org/chainmaker-go/core

go 1.15

require (
	chainmaker.org/chainmaker-go/subscriber v0.0.0
	chainmaker.org/chainmaker/chainconf/v2 v2.1.1-0.20211109075405-cc95de741f5e
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211116072705-04bdd8f42f8d
	chainmaker.org/chainmaker/consensus/v2 v2.0.0-20211117031842-05ab106b7c72
	chainmaker.org/chainmaker/localconf/v2 v2.1.1-0.20211110030026-ce2a7f3760cd
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211110082822-61bdedd084bd
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211116092258-b0de845d438c
	chainmaker.org/chainmaker/txpool-batch/v2 v2.0.0-20211019074609-46e3d29f0908
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20211109074701-81d58330e787
	chainmaker.org/chainmaker/vm v0.0.0-20211028094551-9b2cca96d10d
	github.com/ethereum/go-ethereum v1.10.3 // indirect
	github.com/gogo/protobuf v1.3.2
	github.com/panjf2000/ants/v2 v2.4.3
	github.com/prometheus/client_golang v1.11.0
	github.com/stretchr/testify v1.7.0
)

replace chainmaker.org/chainmaker-go/subscriber => ../subscriber
