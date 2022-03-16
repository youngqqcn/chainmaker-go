/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	_ "net/http/pprof"
	"os"
	"path"
	"strconv"
	"time"

	"chainmaker.org/chainmaker/store/v2/conf"
	"github.com/mitchellh/mapstructure"

	"chainmaker.org/chainmaker-go/module/blockchain"
	"chainmaker.org/chainmaker/localconf/v2"
	"github.com/spf13/cobra"
)

func RebuildDbsCMD() *cobra.Command {
	rebuildDbsCmd := &cobra.Command{
		Use:   "rebuild-dbs",
		Short: "RebuildDbs ChainMaker",
		Long:  "RebuildDbs ChainMaker",
		RunE: func(cmd *cobra.Command, _ []string) error {
			initLocalConfig(cmd)
			backupDbs(rebuildChainId)
			rebuildDbsStart()
			fmt.Println("ChainMaker exit")
			return nil
		},
	}
	attachFlags(rebuildDbsCmd, []string{flagNameOfConfigFilepath, flagNameOfChainId})
	return rebuildDbsCmd
}

func backupDbs(chainId string) {
	timeS := strconv.FormatInt(time.Now().UnixNano(), 10)
	localconf.ChainMakerConfig.StorageConfig["back_path"] = timeS
	localconf.ChainMakerConfig.StorageConfig["rebuild_chainId"] = chainId
	config := &conf.StorageConfig{}
	errThenExit(mapstructure.Decode(localconf.ChainMakerConfig.StorageConfig, config))

	if config.BlockDbConfig.Provider != "leveldb" {
		fmt.Println("Unsupported storage type")
		os.Exit(0)
	}
	oldStorePath :=
		path.Join(config.BlockDbConfig.LevelDbConfig["store_path"].(string), chainId) + "-" + timeS
	isExists, s := pathExists(oldStorePath)
	if s != "" {
		fmt.Println(s)
		os.Exit(0)
	}
	if isExists {
		fmt.Printf(
			"back file(%s) is exists!\n",
			oldStorePath,
		)
		os.Exit(0)
	}

	backupDir(config.BlockDbConfig.LevelDbConfig["store_path"].(string), timeS, chainId)
	backupDir(config.StateDbConfig.LevelDbConfig["store_path"].(string), timeS, chainId)
	backupDir(config.ResultDbConfig.LevelDbConfig["store_path"].(string), timeS, chainId)
	backupDir(config.HistoryDbConfig.LevelDbConfig["store_path"].(string), timeS, chainId)

	if config.TxExistDbConfig != nil {
		backupDir(config.TxExistDbConfig.LevelDbConfig["store_path"].(string), timeS, chainId)
	}

	backupDir(config.StorePath, timeS, chainId)
}

func backupDir(oldPath, timeS, chainId string) {
	newPath := oldPath + "-" + timeS
	errThenExit(os.Mkdir(newPath, os.ModePerm))
	errThenExit(os.Rename(path.Join(oldPath, chainId), path.Join(newPath, chainId)))
	errThenExit(os.RemoveAll(path.Join(oldPath, chainId)))
}

func errThenExit(err error) func() {
	return func() {
		if err != nil {
			fmt.Println(err)
			os.Exit(0)
		}
	}
}

// pathExists is used to determine whether a file or folder exists
func pathExists(path string) (bool, string) {
	if path == "" {
		return false, "invalid parameter, the file path cannot be empty"
	}
	_, err := os.Stat(path)
	if err == nil {
		return true, ""
	}
	if os.IsNotExist(err) {
		return false, ""
	}
	return false, err.Error()
}

func rebuildDbsStart() {
	if localconf.ChainMakerConfig.DebugConfig.IsTraceMemoryUsage {
		traceMemoryUsage()
	}

	// init chainmaker server
	chainMakerServer := blockchain.NewChainMakerServer()
	chainId, _ := localconf.ChainMakerConfig.StorageConfig["rebuild_chainId"].(string)
	if err := chainMakerServer.InitForRebuildDbs(chainId); err != nil {
		log.Errorf("chainmaker server init failed, %s", err.Error())
		return
	}

	// init rpc server
	//rpcServer, err := rpcserver.NewRPCServer(chainMakerServer)
	//if err != nil {
	//	log.Errorf("rpc server init failed, %s", err.Error())
	//	return
	//}

	// init monitor server
	//monitorServer := monitor.NewMonitorServer()

	//// p2p callback to validate
	//txpool.RegisterCallback(rpcServer.Gateway().Invoke)

	// new an error channel to receive errors
	errorC := make(chan error, 1)

	// handle exit signal in separate go routines
	go handleExitSignal(errorC)

	// start blockchains in separate go routines
	if err := chainMakerServer.StartForRebuildDbs(); err != nil {
		log.Errorf("chainmaker server startup failed, %s", err.Error())
		return
	}

	// start rpc server and listen in another go routine
	//if err := rpcServer.Start(); err != nil {
	//	errorC <- err
	//}

	// start monitor server and listen in another go routine
	//if err := monitorServer.Start(); err != nil {
	//	errorC <- err
	//}

	if localconf.ChainMakerConfig.PProfConfig.Enabled {
		startPProf()
	}

	//printLogo()

	// listen error signal in main function
	errC := <-errorC
	if errC != nil {
		log.Error("chainmaker encounters error ", errC)
	}
	//rpcServer.Stop()
	chainMakerServer.Stop()
	log.Info("All is stopped!")

}
