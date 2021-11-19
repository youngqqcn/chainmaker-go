module chainmaker.org/chainmaker-go/consensus

go 1.15

require (
	chainmaker.org/chainmaker/chainconf/v2 v2.1.1-0.20211109075405-cc95de741f5e // indirect
	chainmaker.org/chainmaker/common/v2 v2.1.1-0.20211117131805-630800bfd361
	chainmaker.org/chainmaker/consensus-dpos/v2 v2.0.0-20211117025121-106e166cbbe0
	chainmaker.org/chainmaker/consensus-hotstuff/v2 v2.0.0-20211117031314-728e624e06bd
	chainmaker.org/chainmaker/consensus-raft/v2 v2.0.0-20211117025208-de42ca78a5bc
	chainmaker.org/chainmaker/consensus-solo/v2 v2.0.0-20211117025052-3e60f135dd70
	chainmaker.org/chainmaker/consensus-tbft/v2 v2.0.0-20211117025915-6ab8ea182df8
	chainmaker.org/chainmaker/consensus-utils/v2 v2.0.0-20211115084213-42e840e1efee
	chainmaker.org/chainmaker/localconf/v2 v2.1.1-0.20211110030026-ce2a7f3760cd
	chainmaker.org/chainmaker/pb-go/v2 v2.1.1-0.20211117140137-9095a7ab7a69
	chainmaker.org/chainmaker/protocol/v2 v2.1.1-0.20211116092258-b0de845d438c
	chainmaker.org/chainmaker/utils/v2 v2.1.1-0.20211117144316-3f4a940e94f0 // indirect
	github.com/golang/mock v1.6.0
)

replace github.com/spf13/viper => github.com/spf13/viper v1.7.1 //for go1.15 build
