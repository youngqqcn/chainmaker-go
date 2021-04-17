/*
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package tbft

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"chainmaker.org/chainmaker-go/pb/protogo/common"
	"chainmaker.org/chainmaker-go/pb/protogo/config"
	consensuspb "chainmaker.org/chainmaker-go/pb/protogo/consensus"
	netpb "chainmaker.org/chainmaker-go/pb/protogo/net"

	"chainmaker.org/chainmaker-go/chainconf"

	"chainmaker.org/chainmaker-go/localconf"

	"chainmaker.org/chainmaker-go/protocol"
	"go.uber.org/zap"

	"github.com/gogo/protobuf/proto"

	"chainmaker.org/chainmaker-go/common/msgbus"
	"chainmaker.org/chainmaker-go/utils"

	"chainmaker.org/chainmaker-go/logger"
	tbftpb "chainmaker.org/chainmaker-go/pb/protogo/consensus/tbft"
)

var clog *zap.SugaredLogger = zap.S()

var (
	defaultChanCap    = 1000
	nilHash           = []byte("NilHash")
	consensusStateKey = []byte("ConsensusStateKey")
)

const (
	DefaultTimeoutPropose      = 30 * time.Second // Timeout of waitting for a proposal before prevoting nil
	DefaultTimeoutProposeDelta = 1 * time.Second  // Increased time delta of TimeoutPropose between rounds
	DefaultBlocksPerProposer   = int64(1)         // The number of blocks each proposer can propose
	TimeoutPrevote             = 1 * time.Second  // Timeout of waitting for >2/3 prevote
	TimeoutPrevoteDelta        = 1 * time.Second  // Increased time delta of TimeoutPrevote between round
	TimeoutPrecommit           = 1 * time.Second  // Timeout of waitting for >2/3 precommit
	TimeoutPrecommitDelta      = 1 * time.Second  // Increased time delta of TimeoutPrecommit between round
	TimeoutCommit              = 1 * time.Second
)

// mustMarshal marshals protobuf message to byte slice or panic
func mustMarshal(msg proto.Message) []byte {
	data, err := proto.Marshal(msg)
	if err != nil {
		panic(err)
	}
	return data
}

// mustUnmarshal unmarshals from byte slice to protobuf message or panic
func mustUnmarshal(b []byte, msg proto.Message) {
	if err := proto.Unmarshal(b, msg); err != nil {
		panic(err)
	}
}

// ConsensusTBFTImpl is the implementation of TBFT algorithm
// and it implements the ConsensusEngine interface.
type ConsensusTBFTImpl struct {
	sync.RWMutex
	logger      *logger.CMLogger
	chainID     string
	Id          string
	singer      protocol.SigningMember
	ac          protocol.AccessControlProvider
	dbHandle    protocol.DBHandle
	ledgerCache protocol.LedgerCache
	chainConf   protocol.ChainConf
	netService  protocol.NetService
	msgbus      msgbus.MessageBus
	closeC      chan struct{}

	validatorSet *validatorSet

	Height                 int64
	Round                  int32
	Step                   tbftpb.Step
	Proposal               *Proposal // proposal
	VerifingProposal       *Proposal // verifing proposal
	LockedRound            int32
	LockedProposal         *Proposal // locked proposal
	ValidRound             int32
	ValidProposal          *Proposal // valid proposal
	heightRoundVoteSet     *heightRoundVoteSet
	LastHeightRoundVoteSet *heightRoundVoteSet

	gossip        *gossipService
	timeScheduler *timeScheduler
	verifingBlock *common.Block // verifing block

	proposedBlockC chan *common.Block
	verifyResultC  chan *consensuspb.VerifyResult
	blockHeightC   chan int64
	externalMsgC   chan *tbftpb.TBFTMsg
	internalMsgC   chan *tbftpb.TBFTMsg

	TimeoutPropose      time.Duration
	TimeoutProposeDelta time.Duration

	// time metrics
	metrics *heightMetrics
}

// ConsensusTBFTImplConfig contains initialization config for ConsensusTBFTImpl
type ConsensusTBFTImplConfig struct {
	ChainID     string
	Id          string
	Signer      protocol.SigningMember
	Ac          protocol.AccessControlProvider
	DbHandle    protocol.DBHandle
	LedgerCache protocol.LedgerCache
	ChainConf   protocol.ChainConf
	NetService  protocol.NetService
	MsgBus      msgbus.MessageBus
}

// New creates a tbft consensus instance
func New(config ConsensusTBFTImplConfig) (*ConsensusTBFTImpl, error) {
	consensus := &ConsensusTBFTImpl{}
	consensus.logger = logger.GetLoggerByChain(logger.MODULE_CONSENSUS, config.ChainID)
	consensus.logger.Infof("New ConsensusTBFTImpl[%s]", config.Id)
	consensus.chainID = config.ChainID
	consensus.Id = config.Id
	consensus.singer = config.Signer
	consensus.ac = config.Ac
	consensus.dbHandle = config.DbHandle
	consensus.ledgerCache = config.LedgerCache
	consensus.chainConf = config.ChainConf
	consensus.netService = config.NetService
	consensus.msgbus = config.MsgBus
	consensus.closeC = make(chan struct{})

	consensus.proposedBlockC = make(chan *common.Block, defaultChanCap)
	consensus.verifyResultC = make(chan *consensuspb.VerifyResult, defaultChanCap)
	consensus.blockHeightC = make(chan int64, defaultChanCap)
	consensus.externalMsgC = make(chan *tbftpb.TBFTMsg, defaultChanCap)
	consensus.internalMsgC = make(chan *tbftpb.TBFTMsg, defaultChanCap)

	height, err := config.LedgerCache.CurrentHeight()
	if err != nil {
		return nil, err
	}

	validators, err := GetValidatorListFromConfig(consensus.chainConf.ChainConfig())
	if err != nil {
		return nil, err
	}
	consensus.validatorSet = newValidatorSet(consensus.logger, validators, DefaultBlocksPerProposer)

	consensusStateBytes, err := consensus.dbHandle.Get(consensusStateKey)
	if err != nil {
		return nil, err
	}
	consensus.logger.Infof("new ConsensusTBFTImpl with height: %d, consensusStateBytes == nil %v", height, consensusStateBytes == nil)
	if consensusStateBytes == nil {
		consensus.Height = height + 1
		consensus.Round = 0
		consensus.Step = tbftpb.Step_NewHeight

		consensus.heightRoundVoteSet = newHeightRoundVoteSet(
			consensus.logger, consensus.Height, consensus.Round, consensus.validatorSet)
	} else {
		consensusStateProto := new(tbftpb.ConsensusState)
		mustUnmarshal(consensusStateBytes, consensusStateProto)

		consensus.logger.Infof("new ConsensusTBFTImpl with consensusStateProto [%s](%v/%v/%v), height: %v",
			consensusStateProto.Id,
			consensusStateProto.Height,
			consensusStateProto.Round,
			consensusStateProto.Step,
			height)
		consensus.resetFromProto(consensusStateProto)

		if height >= consensus.Height {
			consensus.Height = height + 1
			consensus.Round = 0
			consensus.Step = tbftpb.Step_NewHeight

			consensus.Proposal = nil
			consensus.heightRoundVoteSet = newHeightRoundVoteSet(
				consensus.logger, consensus.Height, consensus.Round, consensus.validatorSet)
		}
	}

	consensus.timeScheduler = NewTimeSheduler(consensus.logger, config.Id)
	consensus.gossip = newGossipService(consensus.logger, consensus)
	consensus.metrics = NewHeightMetrics(consensus.Height)

	return consensus, nil
}

// Start starts the tbft instance with:
// 1. Register to message bus for subscribing topics
// 2. Start background goroutinues for processing events
// 3. Start timeScheduler for processing timeout shedule
func (consensus *ConsensusTBFTImpl) Start() error {
	consensus.msgbus.Register(msgbus.ProposedBlock, consensus)
	consensus.msgbus.Register(msgbus.VerifyResult, consensus)
	consensus.msgbus.Register(msgbus.RecvConsensusMsg, consensus)
	consensus.msgbus.Register(msgbus.BlockInfo, consensus)
	chainconf.RegisterVerifier(consensuspb.ConsensusType_TBFT, consensus)

	consensus.updateChainConfig()
	consensus.metrics.SetEnterNewHeightTime()
	consensus.logger.Infof("start [%s](%d/%d/%s)",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step)
	if consensus.Step == tbftpb.Step_NewHeight {
		consensus.enterNewRound(consensus.Height, 0)
	} else if consensus.Step == tbftpb.Step_Propose {
		consensus.Step = tbftpb.Step_NewRound
		consensus.enterPropose(consensus.Height, consensus.Round)
	} else if consensus.Step == tbftpb.Step_Precommit || consensus.Step == tbftpb.Step_Commit {
		voteSet := consensus.heightRoundVoteSet.precommits(consensus.Round)
		_, ok := voteSet.twoThirdsMajority()
		if ok {
			height := consensus.Height
			round := consensus.Round
			consensus.Step = tbftpb.Step_Precommit
			consensus.enterCommit(height, round)
		}
		consensus.logger.Infof("restart [%s](%d/%d/%s) TwoThirdsMajority: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, ok)
	}

	consensus.timeScheduler.Start()
	consensus.gossip.start()
	go consensus.handle()
	consensus.logger.Infof("start ConsensusTBFTImpl[%s]", consensus.Id)
	return nil
}

func (consensus *ConsensusTBFTImpl) sendProposeState(isProposer bool) {
	consensus.logger.Debugf("[%s](%d/%d/%s) sendProposeState isProposer: %v",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step, isProposer)
	consensus.msgbus.PublishSafe(msgbus.ProposeState, isProposer)
}

// Stop implements the Stop method of ConsensusEngine interface.
func (consensus *ConsensusTBFTImpl) Stop() error {
	consensus.Lock()
	defer consensus.Unlock()

	consensus.logger.Infof("[%s](%d/%d/%s) stopped", consensus.Id, consensus.Height, consensus.Round, consensus.Step)
	consensus.persistState()
	consensus.gossip.stop()
	close(consensus.closeC)
	return nil
}

// 1. when leadership transfer, change consensus state and send singal
// atomic.StoreInt32()
// proposable <- atomic.LoadInt32(consensus.isLeader)

// 2. when receive pre-prepare block, send block to verifyBlockC
// verifyBlockC <- block

// 3. when receive commit block, send block to commitBlockC
// commitBlockC <- block
func (consensus *ConsensusTBFTImpl) OnMessage(message *msgbus.Message) {
	consensus.logger.Debugf("[%s] OnMessage receive topic: %s", consensus.Id, message.Topic)

	switch message.Topic {
	case msgbus.ProposedBlock:
		if block, ok := message.Payload.(*common.Block); ok {
			consensus.proposedBlockC <- block
		}
	case msgbus.VerifyResult:
		if verifyResult, ok := message.Payload.(*consensuspb.VerifyResult); ok {
			consensus.logger.Debugf("[%s] verify result: %s", consensus.Id, verifyResult.Code)
			consensus.verifyResultC <- verifyResult
		}
	case msgbus.RecvConsensusMsg:
		if msg, ok := message.Payload.(*netpb.NetMsg); ok {
			tbftMsg := new(tbftpb.TBFTMsg)
			mustUnmarshal(msg.Payload, tbftMsg)
			consensus.externalMsgC <- tbftMsg
		} else {
			panic(fmt.Errorf("receive message failed, error message type"))
		}
	case msgbus.BlockInfo:
		if blockInfo, ok := message.Payload.(*common.BlockInfo); ok {
			if blockInfo == nil || blockInfo.Block == nil {
				consensus.logger.Errorf("receive message failed, error message BlockInfo = nil")
				return
			}
			consensus.blockHeightC <- blockInfo.Block.Header.BlockHeight
		} else {
			panic(fmt.Errorf("error message type"))
		}
	}
}

func (consensus *ConsensusTBFTImpl) OnQuit() {
	// do nothing
	//panic("implement me")
}

// Verify implements interface of struct Verifier,
// This interface is used to verify the validity of parameters,
// it executes before consensus.
func (consensus *ConsensusTBFTImpl) Verify(consensusType consensuspb.ConsensusType, chainConfig *config.ChainConfig) error {
	consensus.logger.Infof("[%s](%d/%d/%v) verify chain config",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step)
	if consensusType != consensuspb.ConsensusType_TBFT {
		errMsg := fmt.Sprintf("consensus type is not TBFT: %s", consensusType)
		return errors.New(errMsg)
	}
	config := chainConfig.Consensus
	_, _, _, _, err := consensus.extractConsensusConfig(config)
	return err
}

func (consensus *ConsensusTBFTImpl) updateChainConfig() (addedValidators []string, removedValidators []string, err error) {
	consensus.logger.Infof("[%s](%d/%d/%v) update chain config",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step)

	config := consensus.chainConf.ChainConfig().Consensus
	validators, timeoutPropose, timeoutProposeDelta, tbftBlocksPerProposer, err := consensus.extractConsensusConfig(config)
	if err != nil {
		return nil, nil, err
	}

	consensus.logger.Infof("[%s](%d/%d/%v) update chain config, config: %v, TimeoutPropose: %v, TimeoutProposeDelta: %v, validators: %v",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step, config,
		consensus.TimeoutPropose, consensus.TimeoutProposeDelta, validators)

	consensus.TimeoutPropose = timeoutPropose
	consensus.TimeoutProposeDelta = timeoutProposeDelta
	consensus.validatorSet.updateBlocksPerProposer(tbftBlocksPerProposer)
	return consensus.validatorSet.updateValidators(validators)
}

func (consensus *ConsensusTBFTImpl) extractConsensusConfig(config *config.ConsensusConfig) (validators []string, timeoutPropose time.Duration, timeoutProposeDelta time.Duration, tbftBlocksPerProposer int64, err error) {
	timeoutPropose = DefaultTimeoutPropose
	timeoutProposeDelta = DefaultTimeoutProposeDelta
	tbftBlocksPerProposer = int64(1)

	validators, err = GetValidatorListFromConfig(consensus.chainConf.ChainConfig())
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%v) get validator list from config failed: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, validators)
		return
	}

	for _, v := range config.ExtConfig {
		switch v.Key {
		case protocol.TBFT_propose_timeout_key:
			timeoutPropose, err = consensus.extractProposeTimeout(v.Value)
		case protocol.TBFT_propose_delta_timeout_key:
			timeoutProposeDelta, err = consensus.extractProposeTimeoutDelta(v.Value)
		case protocol.TBFT_blocks_per_proposer:
			tbftBlocksPerProposer, err = consensus.extractBlocksPerProposer(v.Value)
		}

		if err != nil {
			return
		}
	}

	return
}

func (consensus *ConsensusTBFTImpl) extractProposeTimeout(value string) (timeoutPropose time.Duration, err error) {
	if timeoutPropose, err = time.ParseDuration(value); err != nil {
		consensus.logger.Infof("[%s](%d/%d/%v) update chain config, TimeoutPropose: %v, TimeoutProposeDelta: %v, parse TimeoutPropose error: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			consensus.TimeoutPropose, consensus.TimeoutProposeDelta, err)
	}
	return
}

func (consensus *ConsensusTBFTImpl) extractProposeTimeoutDelta(value string) (timeoutProposeDelta time.Duration, err error) {
	if timeoutProposeDelta, err = time.ParseDuration(value); err != nil {
		consensus.logger.Infof("[%s](%d/%d/%v) update chain config, TimeoutPropose: %v, TimeoutProposeDelta: %v, parse TimeoutProposeDelta error: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			consensus.TimeoutPropose, consensus.TimeoutProposeDelta, err)
	}
	return
}

func (consensus *ConsensusTBFTImpl) extractBlocksPerProposer(value string) (tbftBlocksPerProposer int64, err error) {
	if tbftBlocksPerProposer, err = strconv.ParseInt(value, 10, 32); err != nil {
		consensus.logger.Infof("[%s](%d/%d/%v) update chain config, parse BlocksPerProposer error: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, err)
		return
	}
	if tbftBlocksPerProposer <= 0 {
		err = errors.New(fmt.Sprintf("invalid TBFT_blocks_per_proposer: %d", tbftBlocksPerProposer))
		return
	}
	return
}

func (consensus *ConsensusTBFTImpl) handle() {
	consensus.logger.Infof("[%s] handle start", consensus.Id)
	defer consensus.logger.Infof("[%s] handle end", consensus.Id)

	loop := true
	for loop {
		select {
		case block := <-consensus.proposedBlockC:
			consensus.handleProposedBlock(block)
		case result := <-consensus.verifyResultC:
			consensus.handleVerifyResult(result)
		case height := <-consensus.blockHeightC:
			consensus.handleBlockHeight(height)
		case msg := <-consensus.externalMsgC:
			consensus.logger.Debugf("[%s] receive from externalMsgC %s, size: %d", consensus.Id, msg.Type, proto.Size(msg))
			consensus.handleConsensusMsg(msg)
		case msg := <-consensus.internalMsgC:
			consensus.logger.Debugf("[%s] receive from internalMsgC %s, size: %d", consensus.Id, msg.Type, proto.Size(msg))
			consensus.handleConsensusMsg(msg)
		case ti := <-consensus.timeScheduler.GetTimeoutC():
			consensus.handleTimeout(ti)
		case <-consensus.closeC:
			loop = false
			break
		}
	}
}

func (consensus *ConsensusTBFTImpl) handleProposedBlock(block *common.Block) {
	consensus.Lock()
	defer consensus.Unlock()

	consensus.logger.Debugf("[%s](%d/%d/%s) receive proposal from core engine (%d/%x/%d), isProposer: %v",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step,
		block.Header.BlockHeight, block.Header.BlockHash, proto.Size(block), consensus.isProposer(consensus.Height, consensus.Round),
	)

	if block.Header.BlockHeight != consensus.Height {
		consensus.logger.Errorf("[%s](%d/%d/%v) handle proposed block failed, receive block from invalid height: %d",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, block.Header.BlockHeight)
		return
	}

	if !consensus.isProposer(consensus.Height, consensus.Round) {
		consensus.logger.Warnf("[%s](%d/%d/%s) receive proposal from core engine (%d/%x), but isProposer: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			block.Header.BlockHeight, block.Header.BlockHash, consensus.isProposer(consensus.Height, consensus.Round),
		)
		return
	}

	if consensus.Step != tbftpb.Step_Propose {
		consensus.logger.Warnf("[%s](%d/%d/%s) receive proposal from core engine (%d/%x), step error",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			block.Header.BlockHeight, block.Header.BlockHash,
		)
		return
	}

	// Add hash and signature to block
	hash, sig, err := utils.SignBlock(consensus.chainConf.ChainConfig().Crypto.Hash, consensus.singer, block)
	if err != nil {
		consensus.logger.Errorf("[%s]sign block failed, %s", consensus.Id, err)
	}
	block.Header.BlockHash = hash[:]
	block.Header.Signature = sig

	// Add proposal
	proposal := NewProposal(consensus.Id, consensus.Height, consensus.Round, -1, block)
	consensus.signProposal(proposal)
	consensus.Proposal = proposal

	// prevote
	consensus.enterPrevote(consensus.Height, consensus.Round)
}

func (consensus *ConsensusTBFTImpl) handleVerifyResult(verifyResult *consensuspb.VerifyResult) {
	consensus.Lock()
	defer consensus.Unlock()

	height := verifyResult.VerifiedBlock.Header.BlockHeight
	hash := verifyResult.VerifiedBlock.Header.BlockHash

	consensus.logger.Infof("[%s](%d/%d/%s) receive verify result (%d/%x) %v",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step,
		height, hash, verifyResult.Code)

	if consensus.VerifingProposal == nil {
		consensus.logger.Errorf("[%s](%d/%d/%s) receive verify result failed, (%d/%x) %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			height, hash, verifyResult.Code,
		)
		return
	}

	if consensus.Height != height ||
		consensus.Step != tbftpb.Step_Propose ||
		!bytes.Equal(consensus.VerifingProposal.Block.Header.BlockHash, hash) {
		consensus.logger.Warnf("[%s](%d/%d/%s) %x receive verify result (%d/%x) error",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, consensus.VerifingProposal.Block.Header.BlockHash,
			height, hash,
		)
		return
	}

	if verifyResult.Code == consensuspb.VerifyResult_FAIL {
		consensus.logger.Warnf("[%s](%d/%d/%s) %x receive verify result (%d/%x) %v failed",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, consensus.VerifingProposal.Block.Header.BlockHash,
			height, hash, verifyResult.Code,
		)
		return
	}

	consensus.Proposal = consensus.VerifingProposal
	consensus.persistState()
	// Prevote
	consensus.enterPrevote(consensus.Height, consensus.Round)
}

func (consensus *ConsensusTBFTImpl) handleBlockHeight(height int64) {
	consensus.Lock()
	defer consensus.Unlock()

	consensus.logger.Infof("[%s](%d/%d/%s) receive block height %d",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step, height)

	// Outdated block height event
	if consensus.Height > height {
		return
	}

	consensus.logger.Infof("[%s](%d/%d/%s) enterNewHeight because receiving block height %d",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step, height)
	consensus.enterNewHeight(height + 1)
}

func (consensus *ConsensusTBFTImpl) procPropose(msg *tbftpb.TBFTMsg) {
	proposalProto := new(tbftpb.Proposal)
	mustUnmarshal(msg.Msg, proposalProto)

	consensus.logger.Debugf("[%s](%d/%d/%s) receive proposal from %s(%d/%d) (%d/%x/%d)",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step,
		proposalProto.Voter, proposalProto.Height, proposalProto.Round,
		proposalProto.Block.Header.BlockHeight, proposalProto.Block.Header.BlockHash, proto.Size(proposalProto.Block),
	)

	if err := consensus.verifyProposal(proposalProto); err != nil {
		consensus.logger.Debugf("[%s](%d/%d/%s) verify proposal error: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			err,
		)
		return
	}

	proposal := NewProposalFromProto(proposalProto)
	if proposal == nil || proposal.Block == nil {
		consensus.logger.Debugf("[%s](%d/%d/%s) receive invalid proposal because nil proposal",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step)
		return
	}

	height := proposal.Block.Header.BlockHeight
	hash := proposal.Block.Header.BlockHash
	consensus.logger.Debugf("[%s](%d/%d/%s) receive propose %s(%d/%d) hash: %x",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step,
		proposal.Voter, proposal.Height, proposal.Round, hash,
	)

	if !consensus.canReceiveProposal(height, proposal.Round) {
		consensus.logger.Debugf("[%s](%d/%d/%s) receive invalid proposal: %s(%d/%d)",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			proposal.Voter, proposal.Height, proposal.Round,
		)
		return
	}

	proposer, _ := consensus.validatorSet.GetProposer(proposal.Height, proposal.Round)
	if proposer != proposal.Voter {
		consensus.logger.Infof("[%s](%d/%d/%s) proposer: %s, receive proposal from incorrect proposal: %s",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, proposer, proposal.Voter)
		return
	}

	if consensus.Proposal != nil {
		if bytes.Equal(consensus.Proposal.Block.Header.BlockHash, proposal.Block.Header.BlockHash) {
			consensus.logger.Infof("[%s](%d/%d/%s) receive duplicate proposal from proposer: %s(%x)",
				consensus.Id, consensus.Height, consensus.Round, consensus.Step, proposal.Voter, proposal.Block.Header.BlockHash)
		} else {
			consensus.logger.Infof("[%s](%d/%d/%s) receive unequal proposal from proposer: %s(%x)",
				consensus.Id, consensus.Height, consensus.Round, consensus.Step, consensus.Proposal.Block.Header.BlockHash,
				proposal.Voter, proposal.Block.Header.BlockHash)
		}
		return
	}

	if consensus.VerifingProposal != nil {
		if bytes.Equal(consensus.VerifingProposal.Block.Header.BlockHash, proposal.Block.Header.BlockHash) {
			consensus.logger.Infof("[%s](%d/%d/%s) receive proposal which is verifying from proposer: %s(%x)",
				consensus.Id, consensus.Height, consensus.Round, consensus.Step, proposal.Voter, proposal.Block.Header.BlockHash)
		} else {
			consensus.logger.Infof("[%s](%d/%d/%s) receive unequal proposal with verifying proposal from proposer: %s(%x)",
				consensus.Id, consensus.Height, consensus.Round, consensus.Step, consensus.VerifingProposal.Block.Header.BlockHash,
				proposal.Voter, proposal.Block.Header.BlockHash)
		}
		return
	}

	consensus.logger.Debugf("[%s](%d/%d/%s) send for verifying block: (%d-%x)",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step, proposal.Block.Header.BlockHeight, proposal.Block.Header.BlockHash)
	consensus.VerifingProposal = proposal
	consensus.msgbus.PublishSafe(msgbus.VerifyBlock, proposal.Block)
}

func (consensus *ConsensusTBFTImpl) canReceiveProposal(height int64, round int32) bool {
	if consensus.Height != height || consensus.Round != round || consensus.Step < tbftpb.Step_Propose {
		consensus.logger.Debugf("[%s](%d/%d/%s) receive invalid proposal: (%d/%d)",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, height, round)
		return false
	}
	return true
}

func (consensus *ConsensusTBFTImpl) procPrevote(msg *tbftpb.TBFTMsg) {
	prevote := new(tbftpb.Vote)
	mustUnmarshal(msg.Msg, prevote)

	consensus.logger.Debugf("[%s](%d/%d/%s) receive prevote %s(%d/%d/%x)",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step,
		prevote.Voter, prevote.Height, prevote.Round, prevote.Hash,
	)

	if prevote.Voter != consensus.Id {
		err := consensus.verifyVote(prevote)
		if err != nil {
			consensus.logger.Errorf("[%s](%d/%d/%s) receive prevote %s(%d/%d/%x), verifyVote failed: %v",
				consensus.Id, consensus.Height, consensus.Round, consensus.Step,
				prevote.Voter, prevote.Height, prevote.Round, prevote.Hash, err,
			)
			return
		}
	}

	if consensus.Height != prevote.Height ||
		consensus.Round > prevote.Round ||
		(consensus.Round == prevote.Round && consensus.Step > tbftpb.Step_Prevote) {
		consensus.logger.Debugf("[%s](%d/%d/%s) receive invalid prevote %s(%d/%d)",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			prevote.Voter, prevote.Height, prevote.Round,
		)
		return
	}

	vote := NewVoteFromProto(prevote)
	err := consensus.addVote(vote)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%s) addVote %s(%d/%d) failed, %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			prevote.Voter, prevote.Height, prevote.Round, err,
		)
		return
	}
}

func (consensus *ConsensusTBFTImpl) procPrecommit(msg *tbftpb.TBFTMsg) {
	precommit := new(tbftpb.Vote)
	mustUnmarshal(msg.Msg, precommit)

	consensus.logger.Debugf("[%s](%d/%d/%s) receive precommit %s(%d/%d)",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step,
		precommit.Voter, precommit.Height, precommit.Round,
	)

	if precommit.Voter != consensus.Id {
		err := consensus.verifyVote(precommit)
		if err != nil {
			consensus.logger.Errorf("[%s](%d/%d/%s) receive precommit %s(%d/%d/%x), verifyVote failed, %v",
				consensus.Id, consensus.Height, consensus.Round, consensus.Step,
				precommit.Voter, precommit.Height, precommit.Round, precommit.Hash, err,
			)
			return
		}
	}

	if consensus.Height != precommit.Height ||
		consensus.Round > precommit.Round ||
		(consensus.Round == precommit.Round && consensus.Step > tbftpb.Step_Precommit) {
		consensus.logger.Debugf("[%s](%d/%d/%s) receive invalid precommit %s(%d/%d)",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			precommit.Voter, precommit.Height, precommit.Round,
		)
		return
	}

	vote := NewVoteFromProto(precommit)
	err := consensus.addVote(vote)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%s) addVote %s(%d/%d) failed, %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			precommit.Voter, precommit.Height, precommit.Round, err,
		)
		return
	}
}

func (consensus *ConsensusTBFTImpl) handleConsensusMsg(msg *tbftpb.TBFTMsg) {
	consensus.Lock()
	defer consensus.Unlock()

	switch msg.Type {
	case tbftpb.TBFTMsgType_propose:
		consensus.procPropose(msg)
	case tbftpb.TBFTMsgType_prevote:
		consensus.procPrevote(msg)
	case tbftpb.TBFTMsgType_precommit:
		consensus.procPrecommit(msg)
	case tbftpb.TBFTMsgType_state:
		// Async is ok
		go consensus.gossip.onRecvState(msg)
	}
}

// handleTimeout handles timeout event
func (consensus *ConsensusTBFTImpl) handleTimeout(ti timeoutInfo) {
	consensus.Lock()
	defer consensus.Unlock()

	consensus.logger.Infof("[%s](%d/%d/%s) handleTimeout ti: %v",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step, ti)
	switch ti.Step {
	case tbftpb.Step_NewRound:
		consensus.enterNewRound(ti.Height, ti.Round)
	case tbftpb.Step_Prevote:
		consensus.enterPrevote(ti.Height, ti.Round)
	}
}

func (consensus *ConsensusTBFTImpl) commitBlock(block *common.Block) {
	consensus.logger.Debugf("[%s] commitBlock to %d-%x", consensus.Id, block.Header.BlockHeight, block.Header.BlockHash)
	//Simulate a malicious node which commit block without notification
	if localconf.ChainMakerConfig.DebugConfig.IsCommitWithoutPublish {
		consensus.logger.Debugf("[%s](%d/%d/%s) switch IsCommitWithoutPublish: %v, commitBlock block(%d/%x)",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			localconf.ChainMakerConfig.DebugConfig.IsCommitWithoutPublish,
			block.Header.BlockHeight, block.Header.BlockHash,
		)
	} else {
		consensus.logger.Debugf("[%s](%d/%d/%s) commitBlock block(%d/%x)",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			block.Header.BlockHeight, block.Header.BlockHash,
		)
		consensus.msgbus.Publish(msgbus.CommitBlock, block)
		//todo  publishEvent
	}

	consensus.persistState()
}

// ProposeTimeout returns timeout to wait for proposing at `round`
func (consensus *ConsensusTBFTImpl) ProposeTimeout(round int32) time.Duration {
	return time.Duration(
		consensus.TimeoutPropose.Nanoseconds()+consensus.TimeoutProposeDelta.Nanoseconds()*int64(round),
	) * time.Nanosecond
}

// PrevoteTimeout returns timeout to wait for prevoting at `round`
func (consensus *ConsensusTBFTImpl) PrevoteTimeout(round int32) time.Duration {
	return time.Duration(
		TimeoutPrevote.Nanoseconds()+TimeoutPrevoteDelta.Nanoseconds()*int64(round),
	) * time.Nanosecond
}

// PrecommitTimeout returns timeout to wait for precommiting at `round`
func (consensus *ConsensusTBFTImpl) PrecommitTimeout(round int32) time.Duration {
	return time.Duration(
		TimeoutPrecommit.Nanoseconds()+TimeoutPrecommitDelta.Nanoseconds()*int64(round),
	) * time.Nanosecond
}

// CommitTimeout returns timeout to wait for precommiting at `round`
func (consensus *ConsensusTBFTImpl) CommitTimeout(round int32) time.Duration {
	return time.Duration(TimeoutCommit.Nanoseconds()*int64(round)) * time.Nanosecond
}

// AddTimeout adds timeout event to timeScheduler
func (consensus *ConsensusTBFTImpl) AddTimeout(duration time.Duration, height int64, round int32, step tbftpb.Step) {
	consensus.timeScheduler.AddTimeoutInfo(timeoutInfo{duration, height, round, step})
}

// addVote adds `vote` to heightVoteSet
func (consensus *ConsensusTBFTImpl) addVote(vote *Vote) error {
	consensus.logger.Debugf("[%s](%d/%d/%s) addVote %v",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step, vote)

	added, err := consensus.heightRoundVoteSet.addVote(vote)
	if !added || err != nil {
		consensus.logger.Infof("[%s](%d/%d/%s) addVote %v, added: %v, err: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, vote, added, err)
		return err
	}

	consensus.persistState()

	switch vote.Type {
	case tbftpb.VoteType_VotePrevote:
		consensus.addPrevoteVote(vote)
	case tbftpb.VoteType_VotePrecommit:
		consensus.addPrecommitVote(vote)
	}

	// Trigger gossip when receive self vote
	if consensus.Id == vote.Voter {
		go consensus.gossip.triggerEvent()
	}
	return nil
}

func (consensus *ConsensusTBFTImpl) addPrevoteVote(vote *Vote) {
	if consensus.Step != tbftpb.Step_Prevote {
		consensus.logger.Infof("[%s](%d/%d/%s) addVote prevote %v at inappropriate step",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, vote)
		return
	}
	voteSet := consensus.heightRoundVoteSet.prevotes(vote.Round)
	hash, ok := voteSet.twoThirdsMajority()
	if !ok {
		consensus.logger.Debugf("[%s](%d/%d/%s) addVote %v without majority",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, vote)

		if consensus.Round == vote.Round && voteSet.hasTwoThirdsAny() {
			consensus.logger.Infof("[%s](%d/%d/%s) addVote %v with hasTwoThirdsAny",
				consensus.Id, consensus.Height, consensus.Round, consensus.Step, vote)
			consensus.enterPrecommit(consensus.Height, consensus.Round)
		}
		return
	}
	// Upon >2/3 prevotes, Step into StepPrecommit
	if consensus.Proposal != nil {
		if !bytes.Equal(hash, consensus.Proposal.Block.Header.BlockHash) {
			consensus.logger.Errorf("[%s](%d/%d/%s) block matched failed, receive valid block: %x, but unmatched with proposal: %x",
				consensus.Id, consensus.Height, consensus.Round, consensus.Step, hash, consensus.Proposal.Block.Header.BlockHash)
		}
		consensus.enterPrecommit(consensus.Height, consensus.Round)
	} else {
		if isNilHash(hash) {
			consensus.enterPrecommit(consensus.Height, consensus.Round)
		} else {
			consensus.logger.Errorf("[%s](%d/%d/%s) add vote failed, receive valid block: %x, but proposal is nil",
				consensus.Id, consensus.Height, consensus.Round, consensus.Step, hash)
		}
	}
}

func (consensus *ConsensusTBFTImpl) addPrecommitVote(vote *Vote) {
	if consensus.Step != tbftpb.Step_Precommit {
		consensus.logger.Infof("[%s](%d/%d/%s) addVote precommit %v at inappropriate step",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, vote)
		return
	}

	voteSet := consensus.heightRoundVoteSet.precommits(vote.Round)
	hash, ok := voteSet.twoThirdsMajority()
	if !ok {
		consensus.logger.Debugf("[%s](%d/%d/%s) addVote %v without majority",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, vote)

		if consensus.Round == vote.Round && voteSet.hasTwoThirdsAny() {
			consensus.logger.Infof("[%s](%d/%d/%s) addVote %v with hasTwoThirdsAny",
				consensus.Id, consensus.Height, consensus.Round, consensus.Step, vote)
			consensus.enterCommit(consensus.Height, consensus.Round)
		}
		return
	}
	// Upon >2/3 precommits, Step into StepCommit
	if consensus.Proposal != nil {
		if isNilHash(hash) || bytes.Equal(hash, consensus.Proposal.Block.Header.BlockHash) {
			consensus.enterCommit(consensus.Height, consensus.Round)
		} else {
			consensus.logger.Errorf("[%s](%d/%d/%s) block matched failed, receive valid block: %x, but unmatched with proposal: %x",
				consensus.Id, consensus.Height, consensus.Round, consensus.Step, hash, consensus.Proposal.Block.Header.BlockHash)
		}
	} else {
		if !isNilHash(hash) {
			consensus.logger.Errorf("[%s](%d/%d/%s) receive valid block: %x, but proposal is nil",
				consensus.Id, consensus.Height, consensus.Round, consensus.Step, hash)
			return
		}
		consensus.enterCommit(consensus.Height, consensus.Round)
	}
}

// enterNewHeight enter `height`
func (consensus *ConsensusTBFTImpl) enterNewHeight(height int64) {
	consensus.logger.Infof("[%s]attempt enter new height to (%d)", consensus.Id, height)
	if consensus.Height >= height {
		consensus.logger.Errorf("[%s](%v/%v/%v) invalid enter invalid new height to (%v)",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, height)
		return
	}
	addedValidators, removedValidators, err := consensus.updateChainConfig()
	if err != nil {
		consensus.logger.Errorf("[%s](%v/%v/%v) update chain config failed: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, err)
	}
	consensus.gossip.addValidators(addedValidators)
	consensus.gossip.removeValidators(removedValidators)
	consensus.Height = height
	consensus.Round = 0
	consensus.Step = tbftpb.Step_NewHeight
	// consensus.LastHeightRoundVoteSet = consensus.heightRoundVoteSet
	consensus.heightRoundVoteSet = newHeightRoundVoteSet(
		consensus.logger, consensus.Height, consensus.Round, consensus.validatorSet)
	consensus.metrics = NewHeightMetrics(consensus.Height)
	consensus.metrics.SetEnterNewHeightTime()
	consensus.enterNewRound(height, 0)
}

// enterNewRound enter `round` at `height`
func (consensus *ConsensusTBFTImpl) enterNewRound(height int64, round int32) {
	consensus.logger.Debugf("[%s] attempt enterNewRound to (%d/%d)", consensus.Id, height, round)
	if consensus.Height > height ||
		consensus.Round > round ||
		(consensus.Round == round && consensus.Step != tbftpb.Step_NewHeight) {
		consensus.logger.Infof("[%s](%v/%v/%v) enter new round invalid(%v/%v)",

			consensus.Id, consensus.Height, consensus.Round, consensus.Step, height, round)
		return
	}
	consensus.Height = height
	consensus.Round = round
	consensus.Step = tbftpb.Step_NewRound
	consensus.Proposal = nil
	consensus.VerifingProposal = nil
	consensus.metrics.SetEnterNewRoundTime(consensus.Round)
	consensus.enterPropose(height, round)
}

func (consensus *ConsensusTBFTImpl) enterPropose(height int64, round int32) {
	consensus.logger.Debugf("[%s] attempt enterPropose to (%d/%d)", consensus.Id, height, round)
	if consensus.Height != height ||
		consensus.Round > round ||
		(consensus.Round == round && consensus.Step != tbftpb.Step_NewRound) {
		consensus.logger.Infof("[%s](%v/%v/%v) enter invalid propose(%v/%v)",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, height, round)
		return
	}

	// Step into Propose
	consensus.Step = tbftpb.Step_Propose
	consensus.persistState()
	consensus.metrics.SetEnterProposalTime(consensus.Round)
	consensus.AddTimeout(consensus.ProposeTimeout(round), height, round, tbftpb.Step_Prevote)

	//Simulate a node which delay when Propose
	if localconf.ChainMakerConfig.DebugConfig.IsProposeDelay {
		consensus.logger.Infof("[%s](%v/%v/%v) switch IsProposeDelay: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			localconf.ChainMakerConfig.DebugConfig.IsProposeDelay)
		time.Sleep(2 * time.Second)
	}

	//Simulate a malicious node which think itself a proposal
	if localconf.ChainMakerConfig.DebugConfig.IsProposeMultiNodeDuplicately {
		consensus.logger.Infof("[%s](%v/%v/%v) switch IsProposeMultiNodeDuplicately: %v, it always propose",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			localconf.ChainMakerConfig.DebugConfig.IsProposeMultiNodeDuplicately)
		consensus.sendProposeState(true)
	}

	if consensus.isProposer(height, round) {
		consensus.sendProposeState(true)
	}

	go consensus.gossip.triggerEvent()
}

// enterPrevote enter `prevote` phase
func (consensus *ConsensusTBFTImpl) enterPrevote(height int64, round int32) {
	if consensus.Height != height ||
		consensus.Round > round ||
		(consensus.Round == round && consensus.Step != tbftpb.Step_Propose) {
		consensus.logger.Infof("[%s](%v/%v/%v) enter invalid prevote(%v/%v)",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, height, round)
		return
	}

	consensus.logger.Infof("[%s](%v/%v/%v) enter prevote(%v/%v)",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step, height, round)

	// Enter StepPrevote
	consensus.Step = tbftpb.Step_Prevote
	consensus.metrics.SetEnterPrevoteTime(consensus.Round)

	// Disable propose
	consensus.sendProposeState(false)

	var hash []byte = nilHash
	if consensus.Proposal != nil {
		hash = consensus.Proposal.Block.Header.BlockHash
	}

	//Simulate a node which send an invalid(hash=NIL) Prevote
	if localconf.ChainMakerConfig.DebugConfig.IsPrevoteInvalid {
		consensus.logger.Infof("[%s](%v/%v/%v) switch IsPrevoteInvalid: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			localconf.ChainMakerConfig.DebugConfig.IsPrevoteInvalid)
		hash = nil
	}

	//Simulate a node which delay when Propose
	if localconf.ChainMakerConfig.DebugConfig.IsPrevoteDelay {
		consensus.logger.Infof("[%s](%v/%v/%v) switch PrevoteDelay: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			localconf.ChainMakerConfig.DebugConfig.IsPrevoteDelay)
		time.Sleep(2 * time.Second)
	}

	// Broadcast prevote
	// prevote := createPrevoteMsg(consensus.Id, consensus.Height, consensus.Round, hash)
	prevote := NewVote(tbftpb.VoteType_VotePrevote, consensus.Id, consensus.Height, consensus.Round, hash)
	if localconf.ChainMakerConfig.DebugConfig.IsPrevoteOldHeight {
		consensus.logger.Infof("[%s](%v/%v/%v) switch IsPrevoteOldHeight: %v, prevote old height: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			localconf.ChainMakerConfig.DebugConfig.IsPrevoteOldHeight, consensus.Height-1)
		prevote = NewVote(tbftpb.VoteType_VotePrevote, consensus.Id, consensus.Height-1, consensus.Round, hash)
	}
	consensus.signVote(prevote)
	prevoteProto := createPrevoteMsg(prevote)

	consensus.internalMsgC <- prevoteProto
}

// enterPrecommit enter `precommit` phase
func (consensus *ConsensusTBFTImpl) enterPrecommit(height int64, round int32) {
	if consensus.Height != height ||
		consensus.Round > round ||
		(consensus.Round == round && consensus.Step != tbftpb.Step_Prevote) {
		consensus.logger.Infof("[%s](%v/%v/%v) enter precommit invalid(%v/%v)",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, height, round)
		return
	}

	consensus.logger.Infof("[%s](%v/%v/%v) enter precommit(%v/%v)",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step, height, round)

	// Enter StepPrecommit
	consensus.Step = tbftpb.Step_Precommit
	consensus.metrics.SetEnterPrecommitTime(consensus.Round)

	voteSet := consensus.heightRoundVoteSet.prevotes(consensus.Round)
	hash, ok := voteSet.twoThirdsMajority()
	if !ok {
		if voteSet.hasTwoThirdsAny() {
			hash = nilHash
			consensus.logger.Infof("[%s](%v/%v/%v) enter precommit to nil because hasTwoThirdsAny",
				consensus.Id, consensus.Height, consensus.Round, consensus.Step)
		} else {
			panic("this should not happen")
		}
	}

	//Simulate a node which send an invalid(hash=NIL) Precommit
	if localconf.ChainMakerConfig.DebugConfig.IsPrecommitInvalid {
		consensus.logger.Infof("[%s](%v/%v/%v) switch IsPrecommitInvalid: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			localconf.ChainMakerConfig.DebugConfig.IsPrecommitInvalid)
		hash = nil
	}

	// Broadcast precommit
	precommit := NewVote(tbftpb.VoteType_VotePrecommit, consensus.Id, consensus.Height, consensus.Round, hash)
	if localconf.ChainMakerConfig.DebugConfig.IsPrecommitOldHeight {
		consensus.logger.Infof("[%s](%d/%d/%v) switch IsPrecommitOldHeight: %v, precommit old height: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			localconf.ChainMakerConfig.DebugConfig.IsPrecommitOldHeight, consensus.Height-1)
		precommit = NewVote(tbftpb.VoteType_VotePrecommit, consensus.Id, consensus.Height-1, consensus.Round, hash)
	}
	consensus.signVote(precommit)
	precommitProto := createPrecommitMsg(precommit)

	//Simulate a node which delay when Precommit
	if localconf.ChainMakerConfig.DebugConfig.IsPrecommitDelay {
		consensus.logger.Infof("[%s](%v/%v/%v) switch IsPrecommitDelay: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			localconf.ChainMakerConfig.DebugConfig.IsPrecommitDelay)
		time.Sleep(2 * time.Second)
	}

	consensus.internalMsgC <- precommitProto
}

// enterCommit enter `Commit` phase
func (consensus *ConsensusTBFTImpl) enterCommit(height int64, round int32) {
	if consensus.Height != height ||
		consensus.Round > round ||
		(consensus.Round == round && consensus.Step != tbftpb.Step_Precommit) {
		consensus.logger.Infof("[%s](%d/%d/%s) enterCommit invalid(%v/%v)",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, height, round)
		return
	}

	consensus.logger.Infof("[%s](%d/%d/%s) enterCommit(%v/%v)",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step, height, round)

	// Enter StepCommit
	consensus.Step = tbftpb.Step_Commit
	consensus.metrics.SetEnterCommitTime(consensus.Round)
	consensus.logger.Infof("[%s] consensus cost: %s", consensus.Id, consensus.metrics.String())

	voteSet := consensus.heightRoundVoteSet.precommits(consensus.Round)
	hash, ok := voteSet.twoThirdsMajority()
	if !isNilHash(hash) && !ok {
		// This should not happen
		panic(fmt.Errorf("[%s]-%x, enter commit failed, without majority", consensus.Id, hash))
	}

	if isNilHash(hash) {
		// consensus.AddTimeout(consensus.CommitTimeout(round), consensus.Height, round+1, tbftpb.Step_NewRound)
		consensus.enterNewRound(consensus.Height, round+1)
	} else {
		// Proposal block hash must be match with precommited block hash
		if bytes.Compare(hash, consensus.Proposal.Block.Header.BlockHash) != 0 {
			// This should not happen
			panic(fmt.Errorf("[%s] block match failed, unmatch precommit hash: %x with proposal hash: %x",
				consensus.Id, hash, consensus.Proposal.Block.Header.BlockHash))
		}

		qc := mustMarshal(voteSet.ToProto())
		if consensus.Proposal.Block.AdditionalData == nil {
			consensus.Proposal.Block.AdditionalData = &common.AdditionalData{
				ExtraData: make(map[string][]byte),
			}
		}
		consensus.Proposal.Block.AdditionalData.ExtraData[protocol.TBFTAddtionalDataKey] = qc

		// Commit block to core engine
		consensus.commitBlock(consensus.Proposal.Block)
	}
}

func isNilHash(hash []byte) bool {
	return hash == nil || len(hash) == 0 || bytes.Equal(hash, nilHash)
}

// isProposer returns true if this node is proposer at `height` and `round`,
// and returns false otherwise
func (consensus *ConsensusTBFTImpl) isProposer(height int64, round int32) bool {
	proposer, _ := consensus.validatorSet.GetProposer(height, round)

	if proposer == consensus.Id {
		return true
	}
	return false
}

func (consensus *ConsensusTBFTImpl) resetFromProto(csProto *tbftpb.ConsensusState) {
	consensus.Height = csProto.Height
	consensus.Round = csProto.Round
	consensus.Step = csProto.Step
	consensus.Proposal = NewProposalFromProto(csProto.Proposal)
	consensus.VerifingProposal = NewProposalFromProto(csProto.VerifingProposal)
	consensus.heightRoundVoteSet = newHeightRoundVoteSetFromProto(consensus.logger, csProto.HeightRoundVoteSet, consensus.validatorSet)
}

func (consensus *ConsensusTBFTImpl) toProto() *tbftpb.ConsensusState {
	csProto := &tbftpb.ConsensusState{
		Id:                 consensus.Id,
		Height:             consensus.Height,
		Round:              consensus.Round,
		Step:               consensus.Step,
		Proposal:           consensus.Proposal.ToProto(),
		VerifingProposal:   consensus.VerifingProposal.ToProto(),
		HeightRoundVoteSet: consensus.heightRoundVoteSet.ToProto(),
	}
	return csProto
}

func (consensus *ConsensusTBFTImpl) ToProto() *tbftpb.ConsensusState {
	consensus.RLock()
	defer consensus.RUnlock()
	msg := proto.Clone(consensus.toProto())
	return msg.(*tbftpb.ConsensusState)
}

func (consensus *ConsensusTBFTImpl) ToGossipStateProto() *tbftpb.GossipState {
	consensus.RLock()
	defer consensus.RUnlock()

	var proposal []byte
	if consensus.Proposal != nil {
		proposal = consensus.Proposal.Block.Header.BlockHash
	}

	var verifingProposal []byte
	if consensus.Proposal != nil {
		verifingProposal = consensus.Proposal.Block.Header.BlockHash
	}

	gossipProto := &tbftpb.GossipState{
		Id:               consensus.Id,
		Height:           consensus.Height,
		Round:            consensus.Round,
		Step:             consensus.Step,
		Proposal:         proposal,
		VerifingProposal: verifingProposal,
		RoundVoteSet:     consensus.heightRoundVoteSet.getRoundVoteSet(consensus.Round).ToProto(),
	}
	msg := proto.Clone(gossipProto)
	return msg.(*tbftpb.GossipState)
}

func (consensus *ConsensusTBFTImpl) signProposal(proposal *Proposal) error {
	proposalBytes := mustMarshal(proposal.ToProto())
	sig, err := consensus.singer.Sign(consensus.chainConf.ChainConfig().Crypto.Hash, proposalBytes)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%v) sign proposal %s(%d/%d)-%x failed: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			proposal.Voter, proposal.Height, proposal.Round, proposal.Block.Header.BlockHash, err)
		return err
	}

	serializeMember, err := consensus.singer.GetSerializedMember(true)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%v) get serialize member failed: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, err)
		return err
	}
	proposal.Endorsement = &common.EndorsementEntry{
		Signer:    serializeMember,
		Signature: sig,
	}
	return nil
}

func (consensus *ConsensusTBFTImpl) signVote(vote *Vote) error {
	voteBytes := mustMarshal(vote.ToProto())
	sig, err := consensus.singer.Sign(consensus.chainConf.ChainConfig().Crypto.Hash, voteBytes)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%v) sign vote %s(%d/%d)-%x failed: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			vote.Voter, vote.Height, vote.Round, vote.Hash, err)
		return err
	}

	serializeMember, err := consensus.singer.GetSerializedMember(true)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%v) get serialize member failed: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, err)
		return err
	}
	vote.Endorsement = &common.EndorsementEntry{
		Signer:    serializeMember,
		Signature: sig,
	}
	return nil
}

func (consensus *ConsensusTBFTImpl) verifyProposal(proposal *tbftpb.Proposal) error {
	// Verified by idmgmt
	proposalCopy := proto.Clone(proposal)
	proposalCopy.(*tbftpb.Proposal).Endorsement = nil
	message := mustMarshal(proposalCopy)
	principal, err := consensus.ac.CreatePrincipal(
		protocol.ResourceNameConsensusNode,
		[]*common.EndorsementEntry{proposal.Endorsement},
		message,
	)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%s) receive proposal new principal failed, %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, err)
		return err
	}

	result, err := consensus.ac.VerifyPrincipal(principal)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%s) receive proposal VerifyPolicy result: %v, error %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, result, err)
		return err
	}

	if !result {
		consensus.logger.Errorf("[%s](%d/%d/%s) receive proposal VerifyPolicy result: %v, error %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, result, err)
		return fmt.Errorf("VerifyPolicy result: %v", result)
	}

	member, err := consensus.ac.NewMemberFromProto(proposal.Endorsement.Signer)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%s) receive proposal new member failed, %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, err)
		return err
	}
	certId := member.GetMemberId()
	uid, err := consensus.netService.GetNodeUidByCertId(certId)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%s) receive proposal GetNodeUidByCertId failed, %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, err)
		return err
	}

	if uid != proposal.Voter {
		consensus.logger.Errorf("[%s](%d/%d/%s) receive proposal failed, uid %s is not equal with voter %s",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			uid, proposal.Voter)
		return fmt.Errorf("unmatch voter, uid: %v, voter: %v", uid, proposal.Voter)
	}
	return nil
}

func (consensus *ConsensusTBFTImpl) verifyVote(voteProto *tbftpb.Vote) error {
	voteProtoCopy := proto.Clone(voteProto)
	vote := voteProtoCopy.(*tbftpb.Vote)
	vote.Endorsement = nil
	message := mustMarshal(vote)

	principal, err := consensus.ac.CreatePrincipal(
		protocol.ResourceNameConsensusNode,
		[]*common.EndorsementEntry{voteProto.Endorsement},
		message,
	)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%s) verifyVote new policy failed %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, err)
		return err
	}

	result, err := consensus.ac.VerifyPrincipal(principal)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%s) verifyVote verify policy failed %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, err)
		return err
	}

	if !result {
		consensus.logger.Errorf("[%s](%d/%d/%s) verifyVote verify policy result: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, result)
		return fmt.Errorf("verifyVote result: %v", result)
	}

	member, err := consensus.ac.NewMemberFromProto(voteProto.Endorsement.Signer)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%s) verifyVote new member failed %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, err)
		return err
	}

	certId := member.GetMemberId()
	uid, err := consensus.netService.GetNodeUidByCertId(certId)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%s) verifyVote certId: %v, GetNodeUidByCertId failed %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, certId, err)
		return err
	}

	if uid != voteProto.Voter {
		consensus.logger.Errorf("[%s](%d/%d/%s) verifyVote failed, uid %s is not equal with voter %s",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step,
			uid, voteProto.Voter)
		return fmt.Errorf("verifyVote failed, unmatch uid: %v with vote: %v", uid, voteProto.Voter)
	}

	return nil
}

func (consensus *ConsensusTBFTImpl) verifyBlockSignatures(block *common.Block) error {
	consensus.logger.Debugf("[%s](%d/%d/%s) VerifyBlockSignatures block (%d-%x)",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step,
		block.Header.BlockHeight, block.Header.BlockHash)

	blockVoteSet, ok := block.AdditionalData.ExtraData[protocol.TBFTAddtionalDataKey]
	if !ok {
		return fmt.Errorf("verify block signature failed, block.AdditionalData.ExtraData[TBFTAddtionalDataKey] not exist")
	}

	voteSetProto := new(tbftpb.VoteSet)
	if err := proto.Unmarshal(blockVoteSet, voteSetProto); err != nil {
		return err
	}

	voteSet := NewVoteSetFromProto(consensus.logger, voteSetProto, consensus.validatorSet)
	hash, ok := voteSet.twoThirdsMajority()
	if !ok {
		// This should not happen
		return fmt.Errorf("voteSet without majority")
	}

	if !bytes.Equal(hash, block.Header.BlockHash) {
		return fmt.Errorf("hash match failed, unmatch QC: %x to block hash: %v", hash, block.Header.BlockHash)
	}

	consensus.logger.Debugf("[%s](%d/%d/%s) VerifyBlockSignatures block (%d-%x) success",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step,
		block.Header.BlockHeight, block.Header.BlockHash)
	return nil
}

func (consensus *ConsensusTBFTImpl) persistState() {
	begin := time.Now()
	consensusStateProto := consensus.toProto()
	consensusStateBytes := mustMarshal(consensusStateProto)
	consensus.logger.Debugf("[%s](%d/%d/%s) persist state length: %v",
		consensus.Id, consensus.Height, consensus.Round, consensus.Step, len(consensusStateBytes))
	err := consensus.dbHandle.Put(consensusStateKey, consensusStateBytes)
	if err != nil {
		consensus.logger.Errorf("[%s](%d/%d/%s) persist failed, persist to db error: %v",
			consensus.Id, consensus.Height, consensus.Round, consensus.Step, err)
	}
	d := time.Since(begin)
	consensus.metrics.AppendPersistStateDuration(consensus.Round, consensus.Step.String(), d)
}
