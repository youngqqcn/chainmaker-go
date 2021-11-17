go 1.15

require (
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211116072705-04bdd8f42f8d
	chainmaker.org/chainmaker/logger/v2 v2.0.1-0.20211015125919-8e5199930ac9
	chainmaker.org/chainmaker/net-common v0.0.7-0.20211109085844-739f0f904b96
	chainmaker.org/chainmaker/net-libp2p v1.0.1-0.20211117114914-36bd17b2302b
	chainmaker.org/chainmaker/net-liquid v0.0.9-0.20211027111702-03077ba23a67
	chainmaker.org/chainmaker/pb-go/v2 v2.1.0
	chainmaker.org/chainmaker/protocol/v2 v2.0.1-0.20211109074216-fc2674ef6e22
	github.com/gogo/protobuf v1.3.2
	github.com/google/uuid v1.3.0 // indirect
	github.com/hashicorp/golang-lru v0.5.5-0.20210104140557-80c98217689d // indirect
	github.com/huin/goupnp v1.0.1-0.20210310174557-0ca763054c88 // indirect
	github.com/stretchr/testify v1.7.0
	go.uber.org/multierr v1.6.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	honnef.co/go/tools v0.1.3 // indirect
)

replace github.com/libp2p/go-libp2p-core => chainmaker.org/chainmaker/libp2p-core v0.0.2

module chainmaker.org/chainmaker-go/module/net
