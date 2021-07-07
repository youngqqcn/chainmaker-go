/*
 * Copyright (C) BABEC. All rights reserved.
 * Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.
 *
 * SPDX-License-Identifier: Apache-2.0
 */

package native

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"chainmaker.org/chainmaker/pb-go/consts"

	"chainmaker.org/chainmaker-go/utils"

	"chainmaker.org/chainmaker-go/logger"
	commonPb "chainmaker.org/chainmaker/pb-go/common"
	"chainmaker.org/chainmaker/protocol"
)

const (
	PrefixContractInfo     = "Contract:"
	PrefixContractByteCode = "ContractByteCode:"
)

var (
	ContractName             = commonPb.ContractName_SYSTEM_CONTRACT_USER_CONTRACT_MANAGE.String()
	errContractExist         = errors.New("contract exist")
	errContractNotExist      = errors.New("contract not exist")
	errContractVersionExist  = errors.New("contract version exist")
	errContractStatusInvalid = errors.New("contract status invalid")
)

type ContractManager struct {
	methods map[string]ContractFunc
	log     *logger.CMLogger
}

func newContractManager(log *logger.CMLogger) *ContractManager {
	return &ContractManager{
		log:     log,
		methods: registerContractManagerMethods(log),
	}
}

func (c *ContractManager) getMethod(methodName string) ContractFunc {
	return c.methods[methodName]
}

func registerContractManagerMethods(log *logger.CMLogger) map[string]ContractFunc {
	methodMap := make(map[string]ContractFunc, 64)
	runtime := &ContractManagerRuntime{log: log}
	methodMap[consts.ContractManager_INIT_CONTRACT.String()] = runtime.installContract
	methodMap[consts.ContractManager_UPGRADE_CONTRACT.String()] = runtime.upgradeContract
	methodMap[consts.ContractManager_FREEZE_CONTRACT.String()] = runtime.freezeContract
	methodMap[consts.ContractManager_UNFREEZE_CONTRACT.String()] = runtime.unfreezeContract
	methodMap[consts.ContractManager_REVOKE_CONTRACT.String()] = runtime.revokeContract
	methodMap[consts.ContractManager_GET_CONTRACT_INFO.String()] = runtime.getContractInfo
	return methodMap
}
func (r *ContractManagerRuntime) getContractInfo(txSimContext protocol.TxSimContext, parameters map[string]string) ([]byte, error) {
	name := parameters[consts.ContractManager_GetContractInfo_ContractName.String()]
	contract, err := r.GetContractInfo(txSimContext, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(contract)
}
func (r *ContractManagerRuntime) getAllContracts(txSimContext protocol.TxSimContext, parameters map[string]string) ([]byte, error) {
	contracts, err := r.GetAllContracts(txSimContext)
	if err != nil {
		return nil, err
	}
	return json.Marshal(contracts)
}
func (r *ContractManagerRuntime) installContract(txSimContext protocol.TxSimContext, parameters map[string]string) ([]byte, error) {
	name := parameters[consts.ContractManager_Install_ContractName.String()]
	version := parameters[consts.ContractManager_Install_Version.String()]
	byteCodeStr := parameters[consts.ContractManager_Install_ByteCode.String()]
	byteCode, err := base64.StdEncoding.DecodeString(byteCodeStr)
	runtime := parameters[consts.ContractManager_Install_RuntimeType.String()]
	runtimeInt := commonPb.RuntimeType_value[runtime]
	runtimeType := commonPb.RuntimeType(runtimeInt)
	contract, err := r.InstallContract(txSimContext, name, version, byteCode, runtimeType, parameters)
	if err != nil {
		return nil, err
	}
	return contract.Marshal()
}
func (r *ContractManagerRuntime) upgradeContract(txSimContext protocol.TxSimContext, parameters map[string]string) ([]byte, error) {
	name := parameters[consts.ContractManager_Upgrade_ContractName.String()]
	version := parameters[consts.ContractManager_Upgrade_Version.String()]
	byteCodeStr := parameters[consts.ContractManager_Upgrade_ByteCode.String()]
	byteCode, err := base64.StdEncoding.DecodeString(byteCodeStr)
	runtime := parameters[consts.ContractManager_Upgrade_RuntimeType.String()]
	runtimeInt := commonPb.RuntimeType_value[runtime]
	runtimeType := commonPb.RuntimeType(runtimeInt)
	contract, err := r.UpgradeContract(txSimContext, name, version, byteCode, runtimeType, parameters)
	if err != nil {
		return nil, err
	}
	return contract.Marshal()
}
func (r *ContractManagerRuntime) freezeContract(txSimContext protocol.TxSimContext, parameters map[string]string) ([]byte, error) {
	name := parameters[consts.ContractManager_GetContractInfo_ContractName.String()]
	contract, err := r.FreezeContract(txSimContext, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(contract)
}
func (r *ContractManagerRuntime) unfreezeContract(txSimContext protocol.TxSimContext, parameters map[string]string) ([]byte, error) {
	name := parameters[consts.ContractManager_GetContractInfo_ContractName.String()]
	contract, err := r.UnfreezeContract(txSimContext, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(contract)
}
func (r *ContractManagerRuntime) revokeContract(txSimContext protocol.TxSimContext, parameters map[string]string) ([]byte, error) {
	name := parameters[consts.ContractManager_GetContractInfo_ContractName.String()]
	contract, err := r.RevokeContract(txSimContext, name)
	if err != nil {
		return nil, err
	}
	return json.Marshal(contract)
}

type ContractManagerRuntime struct {
	log *logger.CMLogger
}

//GetContractInfo 根据合约名字查询合约的详细信息
func (r *ContractManagerRuntime) GetContractInfo(context protocol.TxSimContext, name string) (*commonPb.Contract, error) {
	if utils.IsAnyBlank(name) {
		err := fmt.Errorf("%s, param[contract_name] of get contract not found", ErrParams.Error())
		r.log.Errorf(err.Error())
		return nil, err
	}
	return utils.GetContractByName(context.Get, name)
}
func (r *ContractManagerRuntime) GetContractByteCode(context protocol.TxSimContext, name string) ([]byte, error) {
	if utils.IsAnyBlank(name) {
		err := fmt.Errorf("%s, param[contract_name] of get contract not found", ErrParams.Error())
		r.log.Errorf(err.Error())
		return nil, err
	}
	return utils.GetContractBytecode(context.Get, name)
}

//GetAllContracts 查询所有合约的详细信息
func (r *ContractManagerRuntime) GetAllContracts(context protocol.TxSimContext) ([]*commonPb.Contract, error) {
	keyPrefix := []byte(PrefixContractInfo)
	it, err := context.Select(commonPb.ContractName_SYSTEM_CONTRACT_USER_CONTRACT_MANAGE.String(), keyPrefix, keyPrefix)
	if err != nil {
		return nil, err
	}
	defer it.Release()
	var result []*commonPb.Contract
	for it.Next() {
		contract := &commonPb.Contract{}
		kv, err := it.Value()
		if err != nil {
			return nil, err
		}
		err = contract.Unmarshal(kv.Value)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

//安装新合约
func (r *ContractManagerRuntime) InstallContract(context protocol.TxSimContext, name, version string, byteCode []byte,
	runTime commonPb.RuntimeType, initParameters map[string]string) (*commonPb.Contract, error) {
	key := []byte(PrefixContractInfo + name)
	//check name exist
	existContract, _ := context.Get(ContractName, key)
	if len(existContract) > 0 { //exist
		return nil, errContractExist
	}
	contract := &commonPb.Contract{
		Name:          name,
		Version:       version,
		RuntimeType:   runTime,
		Status:        commonPb.ContractStatus_NORMAL,
		Type:          commonPb.ContractType_PARALLEL,
		EndorsePolicy: nil,
		UpdatePolicy:  nil,
		CreateTime:    0,
		Creator:       context.GetSender(),
		Address:       nil,
	}
	cdata, _ := contract.Marshal()

	context.Put(ContractName, key, cdata)
	byteCodeKey := []byte(PrefixContractByteCode + name)
	context.Put(ContractName, byteCodeKey, byteCode)
	//实例化合约，并init合约，产生读写集
	context.CallContract(contract, protocol.ContractInitMethod, byteCode, initParameters, 0, commonPb.TxType_INVOKE_CONTRACT)
	return contract, nil
}

//升级现有合约
func (r *ContractManagerRuntime) UpgradeContract(context protocol.TxSimContext, name, version string, byteCode []byte,
	runTime commonPb.RuntimeType, upgradeParameters map[string]string) (*commonPb.Contract, error) {
	key := []byte(PrefixContractInfo + name)
	//check name exist
	existContract, _ := context.Get(ContractName, key)
	if len(existContract) == 0 { //not exist
		return nil, errContractNotExist
	}
	contract := &commonPb.Contract{}
	err := contract.Unmarshal(existContract)
	if err != nil {
		return nil, err
	}
	if contract.Version == version {
		return nil, errContractVersionExist
	}
	contract.RuntimeType = runTime
	//update ContractInfo
	cdata, _ := contract.Marshal()
	context.Put(ContractName, key, cdata)
	//update Contract Bytecode
	byteCodeKey := []byte(PrefixContractByteCode + name)
	context.Put(ContractName, byteCodeKey, byteCode)
	//运行新合约的upgrade方法，产生读写集
	context.CallContract(contract, protocol.ContractUpgradeMethod, byteCode, upgradeParameters, 0, commonPb.TxType_INVOKE_CONTRACT)
	return contract, nil
}
func (r *ContractManagerRuntime) FreezeContract(context protocol.TxSimContext, name string) (*commonPb.Contract, error) {
	return r.changeContractStatus(context, name, commonPb.ContractStatus_NORMAL, commonPb.ContractStatus_FROZEN)
}
func (r *ContractManagerRuntime) UnfreezeContract(context protocol.TxSimContext, name string) (*commonPb.Contract, error) {
	return r.changeContractStatus(context, name, commonPb.ContractStatus_FROZEN, commonPb.ContractStatus_NORMAL)
}
func (r *ContractManagerRuntime) RevokeContract(context protocol.TxSimContext, name string) (*commonPb.Contract, error) {
	return r.changeContractStatus(context, name, commonPb.ContractStatus_NORMAL, commonPb.ContractStatus_REVOKED)
}

func (r *ContractManagerRuntime) changeContractStatus(context protocol.TxSimContext, name string, oldStatus, newStatus commonPb.ContractStatus) (*commonPb.Contract, error) {
	if utils.IsAnyBlank(name) {
		err := fmt.Errorf("%s, param[contract_name] of get contract not found", ErrParams.Error())
		r.log.Errorf(err.Error())
		return nil, err
	}
	contract, err := utils.GetContractByName(context.Get, name)
	if err != nil {
		return nil, err
	}
	if contract.Status != oldStatus {
		return nil, errContractStatusInvalid
	}
	contract.Status = newStatus
	key := []byte(PrefixContractInfo + name)
	cdata, _ := contract.Marshal()
	err = context.Put(ContractName, key, cdata)
	if err != nil {
		return nil, err
	}
	return contract, nil
}
