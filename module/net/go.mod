go 1.15

require (
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211116072705-04bdd8f42f8d
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/net-common v0.0.7-0.20211027061022-edac82204207
	chainmaker.org/chainmaker/net-libp2p v0.0.12-0.20211022065939-a67ee538e812
	chainmaker.org/chainmaker/net-liquid v0.0.9-0.20211027111702-03077ba23a67
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211110082822-61bdedd084bd
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211116092258-b0de845d438c
	github.com/gogo/protobuf v1.3.2
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d // indirect
	github.com/huin/goupnp v1.0.1-0.20210310174557-0ca763054c88 // indirect
	github.com/stretchr/testify v1.7.0
)

replace github.com/libp2p/go-libp2p-core => chainmaker.org/chainmaker/libp2p-core v0.0.2

module chainmaker.org/chainmaker-go/module/net
