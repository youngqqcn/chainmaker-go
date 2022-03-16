package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/common/v2/crypto"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/pb-go/v2/syscontract"
	"chainmaker.org/chainmaker/protocol/v2"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	sdkutils "chainmaker.org/chainmaker/sdk-go/v2/utils"

	"github.com/gogo/protobuf/proto"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

type ParamMultiSign struct {
	Key    string
	Value  string
	IsFile bool
}

func systemContractMultiSignCMD() *cobra.Command {
	systemContractMultiSignCmd := &cobra.Command{
		Use:   "multi-sign",
		Short: "system contract multi sign command",
		Long:  "system contract multi sign command",
	}

	systemContractMultiSignCmd.AddCommand(multiSignReqCMD())
	systemContractMultiSignCmd.AddCommand(multiSignVoteCMD())
	systemContractMultiSignCmd.AddCommand(multiSignQueryCMD())

	return systemContractMultiSignCmd
}

func multiSignReqCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "req",
		Short: "multi sign req",
		Long:  "multi sign req",
		RunE: func(_ *cobra.Command, _ []string) error {
			return multiSignReq()
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagConcurrency, flagTotalCountPerGoroutine, flagSdkConfPath, flagOrgId, flagChainId,
		flagParams, flagTimeout, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagEnableCertHash,
	})

	cmd.MarkFlagRequired(flagSdkConfPath)
	cmd.MarkFlagRequired(flagParams)

	return cmd
}

func multiSignVoteCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vote",
		Short: "multi sign vote",
		Long:  "multi sign vote",
		RunE: func(_ *cobra.Command, _ []string) error {
			return multiSignVote()
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagConcurrency, flagTotalCountPerGoroutine, flagSdkConfPath, flagOrgId, flagChainId, flagTxId,
		flagTimeout, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagEnableCertHash, flagIsAgree,
		flagAdminCrtFilePaths, flagAdminKeyFilePaths, flagSyncResult, flagAdminOrgIds,
	})

	cmd.MarkFlagRequired(flagSdkConfPath)
	cmd.MarkFlagRequired(flagTxId)

	return cmd
}

func multiSignQueryCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "multi sign query",
		Long:  "multi sign query",
		RunE: func(_ *cobra.Command, _ []string) error {
			return multiSignQuery()
		},
	}

	attachFlags(cmd, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath,
		flagConcurrency, flagTotalCountPerGoroutine, flagSdkConfPath, flagOrgId, flagChainId,
		flagTimeout, flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagEnableCertHash, flagTxId,
	})

	cmd.MarkFlagRequired(flagSdkConfPath)
	cmd.MarkFlagRequired(flagTxId)

	return cmd
}

func multiSignReq() error {
	var (
		err     error
		output  []byte
		payload *common.Payload
		client  *sdk.ChainClient
		resp    *common.TxResponse
	)

	client, err = util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
		userSignCrtFilePath, userSignKeyFilePath)
	if err != nil {
		return err
	}
	defer client.Stop()
	var pms []*ParamMultiSign
	var pairs []*common.KeyValuePair
	if params != "" {
		err = json.Unmarshal([]byte(params), &pms)
		if err != nil {
			return err
		}
	}
	for _, pm := range pms {
		if pm.IsFile {
			byteCode, err := ioutil.ReadFile(pm.Value)
			if err != nil {
				panic(err)
			}
			pairs = append(pairs, &common.KeyValuePair{
				Key:   pm.Key,
				Value: byteCode,
			})
		} else {
			pairs = append(pairs, &common.KeyValuePair{
				Key:   pm.Key,
				Value: []byte(pm.Value),
			})
		}

	}
	payload = client.CreateMultiSignReqPayload(pairs)

	resp, err = client.MultiSignContractReq(payload)
	if err != nil {
		return fmt.Errorf("multi sign req failed, %s", err.Error())
	}
	output, err = prettyjson.Marshal(resp)
	if err != nil {
		return err
	}
	fmt.Printf("multi sign req resp: %s\n", string(output))
	return nil
}

func multiSignVote() error {
	var (
		adminKey  string
		adminCrt  string
		adminOrg  string
		adminKeys []string
		adminCrts []string
		adminOrgs []string
		output    []byte
		err       error
		payload   *common.Payload
		endorser  *common.EndorsementEntry
		client    *sdk.ChainClient
		resp      *common.TxResponse
		tx        *common.TransactionInfo
	)

	client, err = util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
		userSignCrtFilePath, userSignKeyFilePath)
	if err != nil {
		return err
	}
	defer client.Stop()

	if sdk.AuthTypeToStringMap[client.GetAuthType()] == protocol.PermissionedWithCert {
		if adminKeyFilePaths != "" {
			adminKeys = strings.Split(adminKeyFilePaths, ",")
		}
		if adminCrtFilePaths != "" {
			adminCrts = strings.Split(adminCrtFilePaths, ",")
		}
		if len(adminKeys) != len(adminCrts) {
			return fmt.Errorf(ADMIN_ORGID_KEY_CERT_LENGTH_NOT_EQUAL_FORMAT, len(adminKeys), len(adminCrts))
		}
		adminKey = adminKeys[0]
		adminCrt = adminCrts[0]
	} else if sdk.AuthTypeToStringMap[client.GetAuthType()] == protocol.PermissionedWithKey {
		if adminKeyFilePaths != "" {
			adminKeys = strings.Split(adminKeyFilePaths, ",")
		}
		if adminOrgIds != "" {
			adminOrgs = strings.Split(adminOrgIds, ",")
		}
		if len(adminKeys) != len(adminOrgs) {
			return fmt.Errorf(ADMIN_ORGID_KEY_LENGTH_NOT_EQUAL_FORMAT, len(adminKeys), len(adminOrgs))
		}
		adminKey = adminKeys[0]
		adminOrg = adminOrgs[0]
	} else {
		if adminKeyFilePaths != "" {
			adminKeys = strings.Split(adminKeyFilePaths, ",")
		}
		if len(adminKeys) == 0 {
			return fmt.Errorf(ADMIN_ORGID_KEY_LENGTH_NOT_EQUAL_FORMAT, len(adminKeys), len(adminOrgs))
		}
		adminKey = adminKeys[0]
	}

	tx, err = client.GetTxByTxId(txId)
	if err != nil {
		return fmt.Errorf("get tx by txid failed, %s", err.Error())
	}
	payload = tx.Transaction.Payload
	if sdk.AuthTypeToStringMap[client.GetAuthType()] == protocol.PermissionedWithCert {
		endorser, err = sdkutils.MakeEndorserWithPath(adminKey, adminCrt, payload)
		if err != nil {
			return fmt.Errorf("multi sign vote failed, %s", err.Error())
		}
	} else if sdk.AuthTypeToStringMap[client.GetAuthType()] == protocol.PermissionedWithKey {
		endorser, err = sdkutils.MakePkEndorserWithPath(adminKey, crypto.HashAlgoMap[client.GetHashType()],
			adminOrg, payload)
		if err != nil {
			return fmt.Errorf("multi sign vote failed, %s", err.Error())
		}
	} else {
		endorser, err = sdkutils.MakePkEndorserWithPath(adminKey, crypto.HashAlgoMap[client.GetHashType()],
			"", payload)
		if err != nil {
			return fmt.Errorf("multi sign vote failed, %s", err.Error())
		}

	}

	resp, err = client.MultiSignContractVote(payload, endorser, isAgree)
	if err != nil {
		return fmt.Errorf("multi sign vote failed, %s", err.Error())
	}
	output, err = prettyjson.Marshal(resp)
	if err != nil {
		return err
	}
	fmt.Printf("multi sign vote resp: %s\n", string(output))

	return nil
}

func multiSignQuery() error {
	var (
		err    error
		resp   *common.TxResponse
		client *sdk.ChainClient
		output []byte
	)

	client, err = util.CreateChainClient(sdkConfPath, chainId, orgId, userTlsCrtFilePath, userTlsKeyFilePath,
		userSignCrtFilePath, userSignKeyFilePath)
	if err != nil {
		return err
	}
	defer client.Stop()

	resp, err = client.MultiSignContractQuery(txId)
	if err != nil {
		return fmt.Errorf("multi sign query failed, %s", err.Error())
	}

	if resp.Code == 0 && resp.ContractResult.Code == 0 {
		multiSignInfo := &syscontract.MultiSignInfo{}
		err = proto.Unmarshal(resp.ContractResult.Result, multiSignInfo)
		if err != nil {
			return err
		}
		output, err = prettyjson.Marshal(multiSignInfo)
		if err != nil {
			return err
		}
		fmt.Printf("multi sign query resp: %s\n", string(output))
		return nil
	}

	output, err = prettyjson.Marshal(resp)
	if err != nil {
		return err
	}
	fmt.Printf("multi sign query resp: %s\n", string(output))
	return nil
}
