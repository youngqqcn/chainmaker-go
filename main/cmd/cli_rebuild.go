/*
Copyright (C) BABEC. All rights reserved.

SPDX-License-Identifier: Apache-2.0
*/

package cmd

import (
	"fmt"
	_ "net/http/pprof"

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
			rebuildDbsStart()
			fmt.Println("ChainMaker exit")
			return nil
		},
	}
	attachFlags(rebuildDbsCmd, []string{flagNameOfConfigFilepath})
	return rebuildDbsCmd
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
