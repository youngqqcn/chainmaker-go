/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

// verify and run contract
package vm

import (
	"chainmaker.org/chainmaker-go/wxvm/xvm"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"chainmaker.org/chainmaker-go/gasm"
	"chainmaker.org/chainmaker-go/logger"
	acPb "chainmaker.org/chainmaker-go/pb/protogo/accesscontrol"
	commonPb "chainmaker.org/chainmaker-go/pb/protogo/common"
	"chainmaker.org/chainmaker-go/protocol"
	"chainmaker.org/chainmaker-go/vm/native"
	"chainmaker.org/chainmaker-go/wasmer"
	"chainmaker.org/chainmaker-go/wxvm"
	"github.com/gogo/protobuf/proto"
)

const WxvmCodeFolder = "wxvm"

type Factory struct {
}

// NewVmManager get vm runtime manager
func (f *Factory) NewVmManager(wxvmCodePathPrefix string, snapshotManager protocol.SnapshotManager, chainId string, AccessControl protocol.AccessControlProvider,
	provider protocol.ChainNodesInfoProvider) protocol.VmManager {

	log := logger.GetLoggerByChain(logger.MODULE_VM, chainId)

	wxvmCodeDir := filepath.Join(wxvmCodePathPrefix, chainId, WxvmCodeFolder)
	log.Infof("init wxvm code dir %s", wxvmCodeDir)
	wasmerVmPoolManager := wasmer.NewVmPoolManager(chainId)
	wxvmCodeManager := xvm.NewCodeManager(chainId, wxvmCodeDir)
	wxvmContextService := xvm.NewContextService(chainId)

	return &ManagerImpl{
		ChainId:                chainId,
		SnapshotManager:        snapshotManager,
		WasmerVmPoolManager:    wasmerVmPoolManager,
		WxvmCodeManager:        wxvmCodeManager,
		WxvmContextService:     wxvmContextService,
		AccessControl:          AccessControl,
		ChainNodesInfoProvider: provider,
		Log:                    log,
	}
}

// Interface of smart contract engine runtime
type RuntimeInstance interface {
	// start vm runtime with invoke, call “method”
	Invoke(contractId *commonPb.ContractId, method string, byteCode []byte, parameters map[string]string,
		txContext protocol.TxSimContext, gasUsed uint64) *commonPb.ContractResult
}

type ManagerImpl struct {
	WasmerVmPoolManager    *wasmer.VmPoolManager
	WxvmCodeManager        *xvm.CodeManager
	WxvmContextService     *xvm.ContextService
	SnapshotManager        protocol.SnapshotManager
	AccessControl          protocol.AccessControlProvider
	ChainNodesInfoProvider protocol.ChainNodesInfoProvider
	ChainId                string
	Log                    *logger.CMLogger
}

func (m *ManagerImpl) GetAccessControl() protocol.AccessControlProvider {
	return m.AccessControl
}

func (m *ManagerImpl) GetChainNodesInfoProvider() protocol.ChainNodesInfoProvider {
	return m.ChainNodesInfoProvider
}

func (m *ManagerImpl) RunContract(contractId *commonPb.ContractId, method string, byteCode []byte, parameters map[string]string,
	txContext protocol.TxSimContext, gasUsed uint64, refTxType commonPb.TxType) (*commonPb.ContractResult, commonPb.TxStatusCode) {

	contractResult := &commonPb.ContractResult{
		Code:    commonPb.ContractResultCode_FAIL,
		Result:  nil,
		Message: "",
	}

	contractName := contractId.ContractName
	if contractName == "" {
		contractResult.Message = "contractName not found"
		return contractResult, commonPb.TxStatusCode_INVALID_CONTRACT_PARAMETER_CONTRACT_NAME
	}

	if parameters == nil {
		parameters = make(map[string]string)
	}

	// return error if contract has been revoked
	revokeKey := []byte(protocol.ContractRevoke)
	if revokeInfo, err := txContext.Get(contractName, revokeKey); err != nil {
		contractResult.Message = fmt.Sprintf("unable to find revoke info for contract:%s,  error:%s", contractName, err.Error())
		return contractResult, commonPb.TxStatusCode_GET_FROM_TX_CONTEXT_FAILED
	} else if len(revokeInfo) != 0 {
		contractResult.Message = fmt.Sprintf("failed to run user contract, %s has been revoked.", contractName)
		return contractResult, commonPb.TxStatusCode_CONTRACT_REVOKE_FAILED
	}

	if native.IsNative(contractName, refTxType) {
		if method == "" {
			contractResult.Message = "require param method not found."
			return contractResult, commonPb.TxStatusCode_INVALID_CONTRACT_PARAMETER_METHOD
		}
		return m.runNativeContract(contractId, method, parameters, txContext)
	} else if m.isUserContract(refTxType) {
		return m.runUserContract(contractId, method, byteCode, parameters, txContext, gasUsed, refTxType)
	} else {
		contractResult.Message = fmt.Sprintf("bad contract call %s, transaction type %s", contractName, refTxType)
		return contractResult, commonPb.TxStatusCode_INVALID_CONTRACT_TRANSACTION_TYPE
	}
}

// runNativeContract invoke native contract
func (m *ManagerImpl) runNativeContract(contractId *commonPb.ContractId, method string, parameters map[string]string,
	txContext protocol.TxSimContext) (*commonPb.ContractResult, commonPb.TxStatusCode) {

	runtimeInstance := native.GetRuntimeInstance(m.ChainId)
	runtimeContractResult := runtimeInstance.Invoke(contractId, method, nil, parameters, txContext)

	if runtimeContractResult.Code == commonPb.ContractResultCode_OK {
		return runtimeContractResult, commonPb.TxStatusCode_SUCCESS
	} else {
		return runtimeContractResult, commonPb.TxStatusCode_CONTRACT_FAIL
	}
}

// runUserContract invoke user contract
func (m *ManagerImpl) runUserContract(contractId *commonPb.ContractId, method string, byteCode []byte, parameters map[string]string,
	txContext protocol.TxSimContext, gasUsed uint64, refTxType commonPb.TxType) (*commonPb.ContractResult, commonPb.TxStatusCode) {

	contractName := contractId.ContractName

	var runtimeType int
	var version = contractId.ContractVersion

	versionKey := []byte(protocol.ContractVersion)
	creatorKey := []byte(protocol.ContractCreator)
	freezeKey := []byte(protocol.ContractFreeze)
	revokeKey := []byte(protocol.ContractRevoke)
	runtimeTypeKey := []byte(protocol.ContractRuntimeType)

	contractResult := &commonPb.ContractResult{Code: commonPb.ContractResultCode_FAIL}

	// return msg if contract has been frozen
	if refTxType == commonPb.TxType_MANAGE_USER_CONTRACT &&
		(method == commonPb.ManageUserContractFunction_UNFREEZE_CONTRACT.String() ||
			method == commonPb.ManageUserContractFunction_REVOKE_CONTRACT.String()) {
		// nothing
	} else if freezeInfo, err := txContext.Get(contractName, freezeKey); err != nil {
		contractResult.Message = fmt.Sprintf("unable to find freeze info for contract:%s,  error:%s", contractName, err.Error())
		return contractResult, commonPb.TxStatusCode_GET_FROM_TX_CONTEXT_FAILED
	} else if len(freezeInfo) != 0 {
		contractResult.Message = fmt.Sprintf("failed to run user contract, %s has been frozen.", contractName)
		return contractResult, commonPb.TxStatusCode_CONTRACT_FREEZE_FAILED
	}

	// init call user contract data
	if refTxType == commonPb.TxType_INVOKE_USER_CONTRACT || refTxType == commonPb.TxType_QUERY_USER_CONTRACT {
		excludeMethodList := make([]string, 3)
		excludeMethodList = append(excludeMethodList, protocol.ContractInitMethod)
		excludeMethodList = append(excludeMethodList, protocol.ContractUpgradeMethod)
		excludeMethodList = append(excludeMethodList, "")

		vt := &verifyType{
			requireVersion:       true,
			requireExcludeMethod: true,
			requireByteCode:      true,
			requireRuntimeType:   true,
			excludeMethodList:    excludeMethodList,
			currentMethod:        method,
		}

		result, code, byteCodeTmp, runtimeTypeTmp := vt.commonVerify(txContext, contractId, contractResult)
		if code != commonPb.TxStatusCode_SUCCESS {
			return result, code
		}

		version = contractId.ContractVersion
		byteCode = byteCodeTmp
		runtimeType = runtimeTypeTmp
	}

	// manager contract logic
	switch method {
	case commonPb.ManageUserContractFunction_INIT_CONTRACT.String():
		method = protocol.ContractInitMethod
		vt := &verifyType{
			requireVersion:       false,
			requireNullVersion:   true,
			requireExcludeMethod: false,
			currentMethod:        method,
			requireFormatVersion: true,
		}
		result, code, _, _ := vt.commonVerify(txContext, contractId, contractResult)
		if code != commonPb.TxStatusCode_SUCCESS {
			return result, code
		}

		// If you call the constructor, you need to take byteCode in the parameter
		if byteCode == nil {
			contractResult.Message = fmt.Sprintf("please provide the bytecode of the contract:%+v while creating contract", contractId)
			return contractResult, commonPb.TxStatusCode_INVALID_CONTRACT_PARAMETER_BYTE_CODE
		}

		if contractId.RuntimeType != commonPb.RuntimeType_INVALID {
			runtimeType = int(contractId.RuntimeType)
		} else {
			contractResult.Message = fmt.Sprintf("please provide the runtime type of the contract:%+v while creating contract", contractId)
			return contractResult, commonPb.TxStatusCode_INVALID_CONTRACT_PARAMETER_RUNTIME_TYPE
		}

		versionedByteCodeKey := append([]byte(protocol.ContractByteCode), []byte(version)...) // <contract name>:B:<contract version>
		// save versioned byteCode
		if err := txContext.Put(contractName, versionKey, []byte(version)); err != nil {
			contractResult.Message = fmt.Sprintf("failed to store byte code for contract:%s, error:%s", contractName, err.Error())
			return contractResult, commonPb.TxStatusCode_PUT_INTO_TX_CONTEXT_FAILED
		}

		// save versioned byteCode
		if err := txContext.Put(contractName, versionedByteCodeKey, byteCode); err != nil {
			contractResult.Message = fmt.Sprintf("failed to store byte code for contract:%s, error:%s", contractName, err.Error())
			return contractResult, commonPb.TxStatusCode_PUT_INTO_TX_CONTEXT_FAILED
		}

		// save sender
		if senderByte, err := proto.Marshal(txContext.GetTx().Header.Sender); err != nil {
			contractResult.Message = fmt.Sprintf("failed to store creator for contract:%s, error:%s", contractName, err.Error())
			return contractResult, commonPb.TxStatusCode_PUT_INTO_TX_CONTEXT_FAILED
		} else {
			if err := txContext.Put(contractName, creatorKey, senderByte); err != nil {
				contractResult.Message = fmt.Sprintf("failed to store creator for contract:%s, error:%s", contractName, err.Error())
				return contractResult, commonPb.TxStatusCode_PUT_INTO_TX_CONTEXT_FAILED
			}
		}

		// save runtime type
		if err := txContext.Put(contractName, runtimeTypeKey, []byte(strconv.Itoa(runtimeType))); err != nil {
			contractResult.Message = fmt.Sprintf("failed to store runtime contract:%s, error:%s", contractName, err.Error())
			return contractResult, commonPb.TxStatusCode_PUT_INTO_TX_CONTEXT_FAILED
		}

		// save txId
		if txIdsBytes, err := txContext.Get(protocol.ContractTxIdsKey, nil); err != nil {
			contractResult.Message = fmt.Sprintf("failed to get tx ids:%s, error:%s", contractName, err.Error())
			return contractResult, commonPb.TxStatusCode_GET_FROM_TX_CONTEXT_FAILED
		} else {
			var txIds []string
			if len(txIdsBytes) > 0 {
				txIds = strings.Split(string(txIdsBytes), ",")
			}
			txIds = append(txIds, txContext.GetTx().Header.TxId)
			if err = txContext.Put(protocol.ContractTxIdsKey, nil, []byte(strings.Join(txIds, ","))); err != nil {
				contractResult.Message = fmt.Sprintf("failed to store tx ids, contract:%s, error:%s", contractName, err.Error())
				return contractResult, commonPb.TxStatusCode_PUT_INTO_TX_CONTEXT_FAILED
			}
		}
	case commonPb.ManageUserContractFunction_UPGRADE_CONTRACT.String():
		method = protocol.ContractUpgradeMethod
		vt := &verifyType{
			requireVersion:       true,
			requireNullVersion:   false,
			requireExcludeMethod: false,
			requireRuntimeType:   true,
			currentMethod:        method,
			requireFormatVersion: true,
		}
		result, code, _, runtimeTypeTmp := vt.commonVerify(txContext, contractId, contractResult)
		if code != commonPb.TxStatusCode_SUCCESS {
			return result, code
		}
		runtimeType = runtimeTypeTmp

		// If you call the constructor, you need to take byteCode in the parameter
		if byteCode == nil {
			contractResult.Message = fmt.Sprintf("please provide the bytecode of the contract:%+v while upgrading", contractId)
			return contractResult, commonPb.TxStatusCode_INVALID_CONTRACT_PARAMETER_BYTE_CODE
		}

		versionedByteCodeKey := append([]byte(protocol.ContractByteCode), []byte(version)...) // <contract name>:B:<contract version>
		// save versioned byteCode
		if err := txContext.Put(contractName, versionKey, []byte(version)); err != nil {
			contractResult.Message = fmt.Sprintf("failed to store byte code for contract:%s, error:%s", contractName, err.Error())
			return contractResult, commonPb.TxStatusCode_PUT_INTO_TX_CONTEXT_FAILED
		}

		// save versioned byteCode
		if err := txContext.Put(contractName, versionedByteCodeKey, byteCode); err != nil {
			contractResult.Message = fmt.Sprintf("failed to store byte code for contract:%s, error:%s", contractName, err.Error())
			return contractResult, commonPb.TxStatusCode_PUT_INTO_TX_CONTEXT_FAILED
		}

		// save runtime type
		if err := txContext.Put(contractName, runtimeTypeKey, []byte(strconv.Itoa(runtimeType))); err != nil {
			contractResult.Message = fmt.Sprintf("failed to store runtime contract:%s, error:%s", contractName, err.Error())
			return contractResult, commonPb.TxStatusCode_PUT_INTO_TX_CONTEXT_FAILED
		}

		// save txId
		if txIdsBytes, err := txContext.Get(protocol.ContractTxIdsKey, nil); err != nil {
			contractResult.Message = fmt.Sprintf("failed to get tx ids:%s, error:%s", contractName, err.Error())
			return contractResult, commonPb.TxStatusCode_GET_FROM_TX_CONTEXT_FAILED
		} else {
			txIds := strings.Split(string(txIdsBytes), ",")
			txIds = append(txIds, txContext.GetTx().Header.TxId)
			if err = txContext.Put(protocol.ContractTxIdsKey, nil, []byte(strings.Join(txIds, ","))); err != nil {
				contractResult.Message = fmt.Sprintf("failed to store tx ids, contract:%s, error:%s", contractName, err.Error())
				return contractResult, commonPb.TxStatusCode_PUT_INTO_TX_CONTEXT_FAILED
			}
		}
	case commonPb.ManageUserContractFunction_FREEZE_CONTRACT.String():
		vt := &verifyType{requireVersion: true}
		result, code, _, _ := vt.commonVerify(txContext, contractId, contractResult)
		if code != commonPb.TxStatusCode_SUCCESS {
			return result, code
		}

		// add freeze target
		if err := txContext.Put(contractName, freezeKey, []byte(contractName)); err != nil {
			contractResult.Message = fmt.Sprintf("failed to store freeze target for contract:%s, error:%s", contractName, err.Error())
			return contractResult, commonPb.TxStatusCode_PUT_INTO_TX_CONTEXT_FAILED
		}
		m.Log.Infof("contract[%s] freeze finish.", contractName)
		contractResult.Code = commonPb.ContractResultCode_OK
		return contractResult, commonPb.TxStatusCode_SUCCESS
	case commonPb.ManageUserContractFunction_UNFREEZE_CONTRACT.String():
		vt := &verifyType{requireVersion: true}
		result, code, _, _ := vt.commonVerify(txContext, contractId, contractResult)
		if code != commonPb.TxStatusCode_SUCCESS {
			return result, code
		}

		// del freeze target
		if err := txContext.Del(contractName, freezeKey); err != nil {
			contractResult.Message = fmt.Sprintf("failed to store unfreeze target for contract:%s, error:%s", contractName, err.Error())
			return contractResult, commonPb.TxStatusCode_PUT_INTO_TX_CONTEXT_FAILED
		}
		m.Log.Infof("contract[%s] unfreeze finish.", contractName)
		contractResult.Code = commonPb.ContractResultCode_OK
		return contractResult, commonPb.TxStatusCode_SUCCESS
	case commonPb.ManageUserContractFunction_REVOKE_CONTRACT.String():
		vt := &verifyType{requireVersion: true}
		result, code, _, _ := vt.commonVerify(txContext, contractId, contractResult)
		if code != commonPb.TxStatusCode_SUCCESS {
			return result, code
		}

		// add revoke target
		if err := txContext.Put(contractName, revokeKey, []byte(contractName)); err != nil {
			contractResult.Message = fmt.Sprintf("failed to store revoke target for contract:%s, error:%s", contractName, err.Error())
			return contractResult, commonPb.TxStatusCode_PUT_INTO_TX_CONTEXT_FAILED
		}
		m.Log.Infof("contract[%s] revoke finish.", contractName)
		contractResult.Code = commonPb.ContractResultCode_OK
		return contractResult, commonPb.TxStatusCode_SUCCESS
	}
	contractId.ContractVersion = version

	return m.invokeUserContractByRuntime(contractId, method, parameters, txContext, runtimeType, byteCode, gasUsed)
}

type verifyType struct {
	requireVersion       bool     // get contract version from TxSimContext, if not exist then return error message
	requireNullVersion   bool     // get contract version from TxSimContext, if exist then return error message
	requireFormatVersion bool     // get contract version from parameter, if format error  then return error message
	requireExcludeMethod bool     // get contract method from parameter, if `currentMethod`in excludeMethodList then return error message
	requireByteCode      bool     // get contract byteCode from parameter, if not exist then return error message
	requireRuntimeType   bool     // get contract runtimeType from TxSimContext, if not exist then return error message
	currentMethod        string   // for requireExcludeMethod
	excludeMethodList    []string // for requireExcludeMethod
}

// commonVerify verify version、method、byteCode、runtimeType, return (result, code, byteCode, runtimeType)
func (v *verifyType) commonVerify(txContext protocol.TxSimContext, contractId *commonPb.ContractId, contractResult *commonPb.ContractResult) (*commonPb.ContractResult, commonPb.TxStatusCode, []byte, int) {
	contractName := contractId.ContractName
	versionKey := []byte(protocol.ContractVersion)
	version := contractId.ContractVersion
	var byteCode []byte
	runtimeType := 0
	msgPre := "verify fail, "

	if v.requireVersion {
		if versionInContext, err := txContext.Get(contractName, versionKey); err != nil {
			contractResult.Message = fmt.Sprintf("%sunable to find latest version for contract:%s,  error:%s",
				msgPre, contractName, err.Error())
			return v.errorResult(contractResult, commonPb.TxStatusCode_GET_FROM_TX_CONTEXT_FAILED)
		} else if len(versionInContext) == 0 {
			contractResult.Message = fmt.Sprintf("%sthe contract does not exist. contract:%s, version is nil",
				msgPre, contractName)
			return v.errorResult(contractResult, commonPb.TxStatusCode_GET_FROM_TX_CONTEXT_FAILED)
		} else {
			contractId.ContractVersion = string(versionInContext)
			version = string(versionInContext)
		}
	}

	if v.requireNullVersion {
		if versionInContext, err := txContext.Get(contractName, versionKey); err != nil {
			contractResult.Message = fmt.Sprintf("%sunable to find latest version for contract:%s,  error:%s",
				msgPre, contractName, err.Error())
			return v.errorResult(contractResult, commonPb.TxStatusCode_GET_FROM_TX_CONTEXT_FAILED)
		} else if versionInContext != nil {
			contractResult.Message = fmt.Sprintf("%sthe contract already exists. contract:%s, version:%s",
				msgPre, contractName, string(versionInContext))
			return v.errorResult(contractResult, commonPb.TxStatusCode_GET_FROM_TX_CONTEXT_FAILED)
		}
	}

	if v.requireExcludeMethod {
		for i := range v.excludeMethodList {
			if v.currentMethod == v.excludeMethodList[i] {
				contractResult.Message = fmt.Sprintf("%scontract:%s, %s method is not allowed to be called",
					msgPre, contractName, v.excludeMethodList[i])
				return v.errorResult(contractResult, commonPb.TxStatusCode_GET_FROM_TX_CONTEXT_FAILED)
			}
		}
	}

	if v.requireFormatVersion {
		if contractId.ContractVersion != "" {
			if len(version) > protocol.DefaultVersionLen {
				contractResult.Message = fmt.Sprintf("%sversion string of the contract:%+v too long while creating contract, "+
					"should be less than %d", msgPre, contractId, protocol.DefaultVersionLen)
				return v.errorResult(contractResult, commonPb.TxStatusCode_GET_FROM_TX_CONTEXT_FAILED)
			}

			match, err := regexp.MatchString(protocol.DefaultVersionRegex, version)
			if err != nil || !match {
				contractResult.Message = fmt.Sprintf("%sversion string of the contract:%+v invalid while invoke user contract, "+
					"should match [%s]", msgPre, contractId, protocol.DefaultVersionRegex)
				return v.errorResult(contractResult, commonPb.TxStatusCode_GET_FROM_TX_CONTEXT_FAILED)
			}

		} else {
			contractResult.Message = fmt.Sprintf("%splease provide the version of the contract:%+v while creating contract ",
				msgPre, contractId)
			return v.errorResult(contractResult, commonPb.TxStatusCode_GET_FROM_TX_CONTEXT_FAILED)
		}
	}

	if v.requireByteCode {
		versionedByteCodeKey := append([]byte(protocol.ContractByteCode), []byte(version)...)
		if byteCodeInContext, err := txContext.Get(contractName, versionedByteCodeKey); err != nil {
			contractResult.Message = fmt.Sprintf("%sfailed to check byte code in tx context for contract %s, %s",
				msgPre, contractName, err.Error())
			return v.errorResult(contractResult, commonPb.TxStatusCode_GET_FROM_TX_CONTEXT_FAILED)
		} else if len(byteCodeInContext) == 0 {
			contractResult.Message = fmt.Sprintf("%sno contract byte code found in the database. "+
				"please create the contract:%s with correct byte code", msgPre, contractName)
			return v.errorResult(contractResult, commonPb.TxStatusCode_INVALID_CONTRACT_PARAMETER_BYTE_CODE)
		} else {
			byteCode = byteCodeInContext
		}
	}

	if v.requireRuntimeType {
		runtimeTypeKey := []byte(protocol.ContractRuntimeType)
		if runtimeTypeBytes, err := txContext.Get(contractName, runtimeTypeKey); err != nil {
			contractResult.Message = fmt.Sprintf("%sfailed to find runtime type %s, error: %s", msgPre, contractName, err.Error())
			return v.errorResult(contractResult, commonPb.TxStatusCode_GET_FROM_TX_CONTEXT_FAILED)
		} else if runtimeTypeTmp, err := strconv.Atoi(string(runtimeTypeBytes)); err != nil {
			contractResult.Message = fmt.Sprintf("%sfailed to convert runtime type for %s, error: %s. "+
				"please create the contract:%s with correct runtime type", msgPre, contractName, err.Error(), contractName)
			return v.errorResult(contractResult, commonPb.TxStatusCode_INVALID_CONTRACT_PARAMETER_RUNTIME_TYPE)
		} else {
			runtimeType = runtimeTypeTmp
		}
	}

	return nil, commonPb.TxStatusCode_SUCCESS, byteCode, runtimeType
}

func (v *verifyType) errorResult(contractResult *commonPb.ContractResult, code commonPb.TxStatusCode) (*commonPb.ContractResult, commonPb.TxStatusCode, []byte, int) {
	return contractResult, code, nil, 0
}

func (m *ManagerImpl) invokeUserContractByRuntime(contractId *commonPb.ContractId, method string, parameters map[string]string,
	txContext protocol.TxSimContext, runtimeType int, byteCode []byte, gasUsed uint64) (*commonPb.ContractResult, commonPb.TxStatusCode) {

	contractResult := &commonPb.ContractResult{Code: commonPb.ContractResultCode_FAIL}

	var runtimeInstance RuntimeInstance
	var err error
	switch commonPb.RuntimeType(runtimeType) {
	case commonPb.RuntimeType_WASMER:
		runtimeInstance, err = m.WasmerVmPoolManager.NewRuntimeInstance(contractId, txContext, byteCode)
		if err != nil {
			contractResult.Message = fmt.Sprintf("failed to create vm runtime, contract: %s, %s", contractId.ContractName, err.Error())
			return contractResult, commonPb.TxStatusCode_CREATE_RUNTIME_INSTANCE_FAILED
		}
	case commonPb.RuntimeType_GASM:
		runtimeInstance = &gasm.RuntimeInstance{
			ChainId: m.ChainId,
			Log:     m.Log,
		}
	case commonPb.RuntimeType_WXVM:
		runtimeInstance = &wxvm.RuntimeInstance{
			ChainId:     m.ChainId,
			CodeManager: m.WxvmCodeManager,
			CtxService:  m.WxvmContextService,
			Log:         m.Log,
		}
	default:
		contractResult.Message = fmt.Sprintf("no such vm runtime %q", runtimeType)
		return contractResult, commonPb.TxStatusCode_INVALID_CONTRACT_PARAMETER_RUNTIME_TYPE
	}

	sender := txContext.GetSender()
	creator := txContext.GetCreator(contractId.ContractName)

	if creator == nil {
		contractResult.Message = fmt.Sprintf("creator is empty for contract:%s", contractId.ContractName)
		return contractResult, commonPb.TxStatusCode_GET_CREATOR_FAILED
	}

	sender, code := getFullCertMember(sender, txContext)
	if code != commonPb.TxStatusCode_SUCCESS {
		return contractResult, code
	}

	creator, code = getFullCertMember(creator, txContext)
	if code != commonPb.TxStatusCode_SUCCESS {
		return contractResult, code
	}

	// Get three items in the certificate: orgid PK role
	if senderMember, err := m.AccessControl.NewMemberFromProto(sender); err != nil {
		contractResult.Message = fmt.Sprintf("failed to unmarshal sender %q", runtimeType)
		return contractResult, commonPb.TxStatusCode_UNMARSHAL_SENDER_FAILED
	} else {
		parameters[protocol.ContractSenderOrgIdParam] = senderMember.GetOrgId()
		parameters[protocol.ContractSenderRoleParam] = string(senderMember.GetRole()[0])
		parameters[protocol.ContractSenderPkParam] = hex.EncodeToString(senderMember.GetSKI())
	}

	// Get three items in the certificate: orgid PK role
	if creatorMember, err := m.AccessControl.NewMemberFromProto(creator); err != nil {
		contractResult.Message = fmt.Sprintf("failed to unmarshal creator %q", runtimeType)
		return contractResult, commonPb.TxStatusCode_UNMARSHAL_CREATOR_FAILED
	} else {
		parameters[protocol.ContractCreatorOrgIdParam] = creator.OrgId
		parameters[protocol.ContractCreatorRoleParam] = string(creatorMember.GetRole()[0])
		parameters[protocol.ContractCreatorPkParam] = hex.EncodeToString(creatorMember.GetSKI())
	}

	parameters[protocol.ContractTxIdParam] = txContext.GetTx().Header.TxId
	parameters[protocol.ContractBlockHeightParam] = strconv.FormatInt(txContext.GetBlockHeight(), 10)

	// calc the gas used by byte code
	// gasUsed := uint64(GasPerByte * len(byteCode))

	tx := txContext.GetTx()
	//m.Log.Debugf("invoke vm, tx id:%s, tx type:%+v, contractId:%+v, method:%+v, runtime type:%+v, byte code len:%+v, params:%+v, sender:%s",
	//	tx.Header.TxId, tx.Header.TxType, contractId, method, runtimeType, len(byteCode), parameters, hex.EncodeToString(tx.Header.Sender.MemberInfo))
	m.Log.Debugf("invoke vm, tx id:%s, tx type:%+v, contractId:%+v, method:%+v, runtime type:%+v, byte code len:%+v",
		tx.Header.TxId, tx.Header.TxType, contractId, method, runtimeType, len(byteCode))

	runtimeContractResult := runtimeInstance.Invoke(contractId, method, byteCode, parameters, txContext, gasUsed)
	if runtimeContractResult.Code == commonPb.ContractResultCode_OK {
		return runtimeContractResult, commonPb.TxStatusCode_SUCCESS
	} else {
		return runtimeContractResult, commonPb.TxStatusCode_CONTRACT_FAIL
	}
}

func getFullCertMember(sender *acPb.SerializedMember, txContext protocol.TxSimContext) (*acPb.SerializedMember, commonPb.TxStatusCode) {
	// If the certificate in the transaction is hash, the original certificate is retrieved
	if !sender.IsFullCert {
		memberInfoHex := hex.EncodeToString(sender.MemberInfo)
		if fullCertMemberInfo, err := txContext.Get(commonPb.ContractName_SYSTEM_CONTRACT_CERT_MANAGE.String(), []byte(memberInfoHex)); err != nil {
			return nil, commonPb.TxStatusCode_GET_SENDER_CERT_FAILED
		} else {
			sender = &acPb.SerializedMember{
				OrgId:      sender.OrgId,
				MemberInfo: fullCertMemberInfo,
				IsFullCert: true,
			}
		}
	}
	return sender, commonPb.TxStatusCode_SUCCESS
}

func (m *ManagerImpl) isUserContract(refTxType commonPb.TxType) bool {
	switch refTxType {
	case commonPb.TxType_MANAGE_USER_CONTRACT,
		commonPb.TxType_INVOKE_USER_CONTRACT,
		commonPb.TxType_QUERY_USER_CONTRACT:
		return true
	default:
		return false
	}
}
