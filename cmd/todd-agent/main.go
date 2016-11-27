/*
	Primary entry point for ToDD Agent

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd/agent/cache"
	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/agent/facts"
	"github.com/toddproject/todd/comms"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/hostresources"
)

func main() {
	// Command-line Arguments
	configPath := flag.String("config", "/etc/todd/agent.cfg", "ToDD agent config file location")
	flag.Usage = func() {
		fmt.Print(`Usage: todd-agent [OPTIONS] COMMAND [arg...]

    An extensible framework for providing natively distributed testing on demand

    Options:
      --config="/etc/todd/agent.cfg"          Absolute path to ToDD agent config file`, "\n\n")

		os.Exit(0)
	}
	flag.Parse()

	// TODO(moswalt): Implement configurable loglevel in server and agent
	log.SetLevel(log.DebugLevel)

	cfg, err := config.GetConfig(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	// Set up cache
	ac, err := cache.Open(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer ac.Close()

	// Generate UUID
	uuid := hostresources.GenerateUUID()
	err = ac.SetKeyValue("uuid", uuid)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("ToDD Agent Activated: %s", uuid)

	// Construct comms package
	tc, err := comms.NewAgentComms(cfg, ac)
	if err != nil {
		log.Fatal(err)
	}

	// Spawn goroutine to listen for tasks issued by server
	go func() {
		for {
			err := tc.ListenForTasks(uuid)
			if err != nil {
				log.Warn("ListenForTasks reported a failure. Trying again...")
			}
		}
	}()

	// Watch for changes to group membership
	go tc.WatchForGroup()

	// Continually advertise agent status into message queue
	advertiseAgent(cfg, tc, uuid)
}

func advertiseAgent(cfg config.Config, tc comms.Package, uuid string) {
	ticker := time.NewTicker(10 * time.Second) // TODO(moswalt): make configurable
	for {
		// Gather assets here as a map, and refer to a key in that map in the below struct
		factCollectors, testlets, err := getLocalAssets(cfg.LocalResources.OptDir)
		if err != nil {
			log.Error("Error gathering assets:", err)
			<-ticker.C
			continue
		}

		defaultaddr := cfg.LocalResources.IPAddrOverride
		if defaultaddr == "" {
			defaultaddr = hostresources.GetIPOfInt(cfg.LocalResources.DefaultInterface).String()
		}

		fcts, err := facts.GetFacts(cfg)
		if err != nil {
			log.Error("Error gathering facts:", err)
			<-ticker.C
			continue
		}

		// Create an AgentAdvert instance to represent this particular agent
		me := defs.AgentAdvert{
			UUID:           uuid,
			DefaultAddr:    defaultaddr,
			FactCollectors: factCollectors,
			Testlets:       testlets,
			Facts:          fcts,
			LocalTime:      time.Now().UTC(),
		}

		// Advertise this agent
		err = tc.AdvertiseAgent(me)
		if err != nil {
			log.Error("Failed to advertise agent after several retries")
		}
		<-ticker.C
	}
}
