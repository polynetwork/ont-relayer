/*
 * Copyright (C) 2018 The ontology Authors
 * This file is part of The ontology library.
 *
 * The ontology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The ontology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The ontology.  If not, see <http://www.gnu.org/licenses/>.
 */
/*
* Copyright (C) 2020 The poly network Authors
* This file is part of The poly network library.
*
* The poly network is free software: you can redistribute it and/or modify
* it under the terms of the GNU Lesser General Public License as published by
* the Free Software Foundation, either version 3 of the License, or
* (at your option) any later version.
*
* The poly network is distributed in the hope that it will be useful,
* but WITHOUT ANY WARRANTY; without even the implied warranty of
* MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
* GNU Lesser General Public License for more details.
* You should have received a copy of the GNU Lesser General Public License
* along with The poly network . If not, see <http://www.gnu.org/licenses/>.
*/
package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	sdk "github.com/ontio/ontology-go-sdk"
	"github.com/polynetwork/ont-relayer/cmd"
	"github.com/polynetwork/ont-relayer/common"
	"github.com/polynetwork/ont-relayer/config"
	"github.com/polynetwork/ont-relayer/log"
	"github.com/polynetwork/ont-relayer/service"
	asdk "github.com/polynetwork/poly-go-sdk"
	"github.com/urfave/cli"
)

func setupApp() *cli.App {
	app := cli.NewApp()
	app.Usage = "Relayer cli"
	app.Action = startSync
	app.Copyright = "Copyright in 2018 The Ontology Authors"
	app.Flags = []cli.Flag{
		cmd.LogLevelFlag,
		cmd.LogDir,
		cmd.ConfigPathFlag,
		cmd.OntPwd,
		cmd.AlliaPwd,
	}
	app.Commands = []cli.Command{}
	app.Before = func(context *cli.Context) error {
		runtime.GOMAXPROCS(runtime.NumCPU())
		return nil
	}
	return app
}

func main() {
	if err := setupApp().Run(os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startSync(ctx *cli.Context) {
	logLevel := ctx.GlobalInt(cmd.GetFlagName(cmd.LogLevelFlag))
	logDir := ctx.GlobalString(cmd.GetFlagName(cmd.LogDir))
	log.InitLog(logLevel, log.Stdout, logDir)
	configPath := ctx.String(cmd.GetFlagName(cmd.ConfigPathFlag))
	err := config.DefConfig.Init(configPath)
	if err != nil {
		fmt.Println("DefConfig.Init error:", err)
		return
	}

	ontPwd := ctx.GlobalString(cmd.GetFlagName(cmd.OntPwd))
	alliaPwd := ctx.GlobalString(cmd.GetFlagName(cmd.AlliaPwd))

	aliaSdk := asdk.NewPolySdk()
	if err := SetUpPoly(aliaSdk, config.DefConfig.AliaJsonRpcAddress); err != nil {
		panic(fmt.Errorf("failed to set up poly: %v", err))
	}
	sideSdk := sdk.NewOntologySdk()
	sideSdk.NewRpcClient().SetAddress(config.DefConfig.SideJsonRpcAddress)
	aliaAccount, ok := common.GetAliaAccountByPassword(aliaSdk, config.DefConfig.AliaWalletFile, alliaPwd)
	if !ok {
		fmt.Println("common.GetAccountByPassword error")
		return
	}
	sideAccount, ok := common.GetSideAccountByPassword(sideSdk, config.DefConfig.SideWalletFile, ontPwd)
	if !ok {
		fmt.Println("common.GetAccountByPassword error")
		return
	}

	syncService := service.NewSyncService(aliaAccount, sideAccount, aliaSdk, sideSdk)
	syncService.Run()

	waitToExit()
}

func waitToExit() {
	exit := make(chan bool, 0)
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	go func() {
		for sig := range sc {
			log.Infof("Ontology received exit signal:%v.", sig.String())
			close(exit)
			break
		}
	}()
	<-exit
}

func SetUpPoly(poly *asdk.PolySdk, rpcAddr string) error {
	poly.NewRpcClient().SetAddress(rpcAddr)
	hdr, err := poly.GetHeaderByHeight(0)
	if err != nil {
		return err
	}
	poly.SetChainId(hdr.ChainID)
	return nil
}
