go 1.15

require (
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211117131805-630800bfd361
	chainmaker.org/chainmaker/logger/v2 v2.1.1-0.20211109074349-f79af5e1892d
	chainmaker.org/chainmaker/net-common v0.0.7-0.20211109085844-739f0f904b96
	chainmaker.org/chainmaker/net-libp2p v1.0.1-0.20211109090515-4889a63c74af
	chainmaker.org/chainmaker/net-liquid v1.0.1-0.20211109090708-3be221384c3e
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211117140137-9095a7ab7a69
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211119081550-ff85fecc318d
	github.com/gogo/protobuf v1.3.2
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d // indirect
	github.com/huin/goupnp v1.0.1-0.20210310174557-0ca763054c88 // indirect
	github.com/stretchr/testify v1.7.0
)

replace github.com/libp2p/go-libp2p-core => chainmaker.org/chainmaker/libp2p-core v0.0.2

module chainmaker.org/chainmaker-go/module/net
