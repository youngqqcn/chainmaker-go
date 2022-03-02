/*
Copyright (C) BABEC. All rights reserved.
Copyright (C) THL A29 Limited, a Tencent company. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package client

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"chainmaker.org/chainmaker-go/tools/cmc/util"
	"chainmaker.org/chainmaker/pb-go/v2/common"
	"chainmaker.org/chainmaker/protocol/v2"
	sdk "chainmaker.org/chainmaker/sdk-go/v2"
	sdkutils "chainmaker.org/chainmaker/sdk-go/v2/utils"
	"github.com/hokaccha/go-prettyjson"
	"github.com/spf13/cobra"
)

const (
	flagCertAlias   = "cert-alias"
	flagNewCertPath = "new-cert-path"
)

var (
	certAlias   string
	newCertPath string
)

func certAliasCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "certalias",
		Short: "cert alias command",
		Long:  "cert alias command",
	}
	cmd.AddCommand(updateCertByAliasCMD())
	cmd.AddCommand(deleteCertAliasCMD())
	cmd.AddCommand(queryCertAliasCMD())
	return cmd
}

func updateCertByAliasCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "update cert by alias",
		Long:  "update cert by alias",
		RunE: func(_ *cobra.Command, _ []string) error {
			var adminKeys, adminCrts []string

			if adminKeyFilePaths != "" {
				adminKeys = strings.Split(adminKeyFilePaths, ",")
			}
			if adminCrtFilePaths != "" {
				adminCrts = strings.Split(adminCrtFilePaths, ",")
			}
			if len(adminKeys) != len(adminCrts) {
				return fmt.Errorf(ADMIN_ORGID_KEY_CERT_LENGTH_NOT_EQUAL_FORMAT, len(adminKeys), len(adminCrts))
			}

			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			// required PermissionedWithCert mode
			if sdk.AuthTypeToStringMap[cc.GetAuthType()] != protocol.PermissionedWithCert {
				return errors.New("cert alias only for PermissionedWithCert mode")
			}

			//// 2. update cert alias for myself
			newCertPEM, err := ioutil.ReadFile(newCertPath)
			if err != nil {
				return fmt.Errorf("read cert file failed, %s", err)
			}
			payload := cc.CreateUpdateCertByAliasPayload(certAlias, string(newCertPEM))
			endorsementEntrys := make([]*common.EndorsementEntry, len(adminKeys))
			for i := range adminKeys {
				e, err := sdkutils.MakeEndorserWithPath(adminKeys[i], adminCrts[i], payload)
				if err != nil {
					return fmt.Errorf("sign payload failed, %s", err.Error())
				}

				endorsementEntrys[i] = e
			}

			if certAlias == cc.GetLocalCertAlias() {
				syncResult = false
			}

			resp, err := cc.UpdateCertByAlias(payload, endorsementEntrys, -1, syncResult)
			if err != nil {
				return fmt.Errorf("send request failed, %s", err.Error())
			}

			err = util.CheckProposalRequestResp(resp, false)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(resp)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	pFlags := cmd.PersistentFlags()
	pFlags.StringVar(&certAlias, flagCertAlias, "", "cert alias")
	pFlags.StringVar(&newCertPath, flagNewCertPath, "", "new cert file path for update cert by alias")

	util.AttachFlags(cmd, flags, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagOrgId, flagChainId, flagSyncResult,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagAdminCrtFilePaths, flagAdminKeyFilePaths,
		flagEnableCertHash,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	return cmd
}

func deleteCertAliasCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "delete cert alias",
		Long:  "delete cert alias",
		RunE: func(_ *cobra.Command, _ []string) error {
			var adminKeys, adminCrts []string

			if adminKeyFilePaths != "" {
				adminKeys = strings.Split(adminKeyFilePaths, ",")
			}
			if adminCrtFilePaths != "" {
				adminCrts = strings.Split(adminCrtFilePaths, ",")
			}
			if len(adminKeys) != len(adminCrts) {
				return fmt.Errorf(ADMIN_ORGID_KEY_CERT_LENGTH_NOT_EQUAL_FORMAT, len(adminKeys), len(adminCrts))
			}

			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			// required PermissionedWithCert mode
			if sdk.AuthTypeToStringMap[cc.GetAuthType()] != protocol.PermissionedWithCert {
				return errors.New("cert alias only for PermissionedWithCert mode")
			}

			//// 2. delete cert alias
			payload := cc.CreateDeleteCertsAliasPayload([]string{certAlias})
			endorsementEntrys := make([]*common.EndorsementEntry, len(adminKeys))
			for i := range adminKeys {
				e, err := sdkutils.MakeEndorserWithPath(adminKeys[i], adminCrts[i], payload)
				if err != nil {
					return fmt.Errorf("sign payload failed, %s", err.Error())
				}

				endorsementEntrys[i] = e
			}

			if certAlias == cc.GetLocalCertAlias() {
				syncResult = false
			}

			resp, err := cc.DeleteCertsAlias(payload, endorsementEntrys, -1, syncResult)
			if err != nil {
				return fmt.Errorf("send request failed, %s", err.Error())
			}

			err = util.CheckProposalRequestResp(resp, false)
			if err != nil {
				return err
			}

			output, err := prettyjson.Marshal(resp)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	pFlags := cmd.PersistentFlags()
	pFlags.StringVar(&certAlias, flagCertAlias, "", "cert alias")

	util.AttachFlags(cmd, flags, []string{
		flagUserSignKeyFilePath, flagUserSignCrtFilePath, flagOrgId, flagChainId, flagSyncResult,
		flagUserTlsCrtFilePath, flagUserTlsKeyFilePath, flagAdminCrtFilePaths, flagAdminKeyFilePaths,
		flagEnableCertHash,
	})

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	return cmd
}

func queryCertAliasCMD() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "query cert alias",
		Long:  "query cert alias",
		RunE: func(_ *cobra.Command, _ []string) error {
			//// 1.Chain Client
			cc, err := util.CreateChainClientWithConfPath(sdkConfPath, false)
			if err != nil {
				return err
			}
			defer cc.Stop()

			// required PermissionedWithCert mode
			if sdk.AuthTypeToStringMap[cc.GetAuthType()] != protocol.PermissionedWithCert {
				return errors.New("cert alias only for PermissionedWithCert mode")
			}

			//// 2. query cert alias
			aliasInfos, err := cc.QueryCertsAlias([]string{certAlias})
			if err != nil {
				return fmt.Errorf("send request failed, %s", err.Error())
			}

			output, err := prettyjson.Marshal(aliasInfos)
			if err != nil {
				return err
			}
			fmt.Println(string(output))
			return nil
		},
	}

	pFlags := cmd.PersistentFlags()
	pFlags.StringVar(&certAlias, flagCertAlias, "", "cert alias")

	util.AttachAndRequiredFlags(cmd, flags, []string{
		flagSdkConfPath,
	})
	return cmd
}
