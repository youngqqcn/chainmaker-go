/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	_ "net/http/pprof"
	"os"
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
			backupDbs()
			rebuildDbsStart()
			fmt.Println("ChainMaker exit")
			return nil
		},
	}
	attachFlags(rebuildDbsCmd, []string{flagNameOfConfigFilepath})
	return rebuildDbsCmd
}
func backupDbs() {
	timeS := time.Now().String()
	localconf.ChainMakerConfig.StorageConfig["back_path"] = timeS
	config := &conf.StorageConfig{}
	err := mapstructure.Decode(localconf.ChainMakerConfig.StorageConfig, config)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	if config.BlockDbConfig.Provider != "leveldb" {
		fmt.Println("Unsupported storage type")
		os.Exit(0)
	}
	oldStorePath :=
		config.BlockDbConfig.LevelDbConfig["store_path"].(string) + "-" + timeS
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
	err = os.Rename(config.BlockDbConfig.LevelDbConfig["store_path"].(string),
		config.BlockDbConfig.LevelDbConfig["store_path"].(string)+"-"+timeS)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	err = os.Rename(config.StateDbConfig.LevelDbConfig["store_path"].(string),
		config.StateDbConfig.LevelDbConfig["store_path"].(string)+"-"+timeS)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	err = os.Rename(config.ResultDbConfig.LevelDbConfig["store_path"].(string),
		config.ResultDbConfig.LevelDbConfig["store_path"].(string)+"-"+timeS)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	err = os.Rename(config.HistoryDbConfig.LevelDbConfig["store_path"].(string),
		config.HistoryDbConfig.LevelDbConfig["store_path"].(string)+"-"+timeS)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	err = os.Rename(config.TxExistDbConfig.LevelDbConfig["store_path"].(string),
		config.TxExistDbConfig.LevelDbConfig["store_path"].(string)+"-"+timeS)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
	}
	err = os.Rename(config.StorePath,
		config.StorePath+"-"+timeS)
	if err != nil {
		fmt.Println(err)
		os.Exit(0)
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
	if err := chainMakerServer.InitForRebuildDbs(); err != nil {
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
