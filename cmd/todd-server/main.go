/*
    ToDD Server - primary entry point

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd"
	toddapi "github.com/toddproject/todd/api/server"
	"github.com/toddproject/todd/comms"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/db"
	"github.com/toddproject/todd/server"
	"github.com/toddproject/todd/server/grouping"
)

func main() {
	configPath := flag.String("config", "/etc/todd/server.cfg", "ToDD server config file location")
	flag.Usage = func() {
		fmt.Print(`Usage: todd-server [OPTIONS] COMMAND [arg...]

    An extensible framework for providing natively distributed testing on demand

    Options:
      --config="/etc/todd/server.cfg"          Absolute path to ToDD server config file`, "\n\n")

		os.Exit(0)
	}
	flag.Parse()

	// TODO(moswalt): Implement configurable loglevel in server and agent
	log.SetLevel(log.DebugLevel)

	cfg, err := config.GetConfig(*configPath)
	if err != nil {
		log.Fatalf("Problem getting configuration: %v\n", err)
	}

	// Start serving collectors and testlets, and retrieve map of names and hashes
	assets := newAssetConfig(cfg)

	// Perform database initialization tasks
	tdb, err := db.New(cfg)
	if err != nil {
		log.Fatalf("Error setting up database: %v", err)
	}

	// Start listening for agent advertisements
	tc, err := comms.New(cfg)
	if err != nil {
		log.Fatalf("Problem connecting to comms: %v\n", err)
	}

	srv := server.New(cfg, tc, tdb, assets)

	// Initialize API
	tapi := toddapi.ServerAPI{Server: srv}
	go func() {
		log.Fatal(tapi.Start(cfg, tdb))
	}()

	ctx, done := context.WithCancel(context.Background())
	defer done()

	agents, err := tc.ListenForAgent(ctx)
	if err != nil {
		log.Fatal("Error listening for ToDD Agents:", err)
	}
	go func() {
		for msg := range agents {
			srv.HandleAgentAdvertisement(msg)
		}
	}()

	// Kick off group calculation in background
	go func() {
		for {
			log.Info("Beginning group calculation")
			grouping.CalculateGroups(cfg, tdb)
			time.Sleep(time.Second * time.Duration(cfg.Grouping.Interval))
		}
	}()

	log.Infof("ToDD server v%s.", todd.Version)

	select {} // Block without interrupting CPU
}
