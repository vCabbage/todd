/*
   ToDD Agent

   Copyright 2015 - Matt Oswalt
*/

package main

import (
	"time"

	log "github.com/mierdin/todd/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	"github.com/mierdin/todd/agent/defs"
	"github.com/mierdin/todd/comms"
	"github.com/mierdin/todd/config"
	"github.com/mierdin/todd/facts"
)

func init() {
	// TODO(moswalt): Implement configurable loglevel in server and agent
	log.SetLevel(log.DebugLevel)
}

func main() {

	//TODO (moswalt): Need to make this configurable
	cfg := config.GetConfig("/etc/agent_config.cfg")

	// Create an AgentAdvert instance to represent this particular agent
	var me defs.AgentAdvert

	// Generate UUID
	me.Uuid = generateUuid()

	log.Infof("ToDD Agent Activated: %s", me.Uuid)

	// Construct comms package
	var tc = comms.NewToDDComms(cfg)

	// Spawn goroutine to remediate any fact issues when informed by server
	go tc.CommsPackage.ListenForFactRemediationNotice(me.Uuid)

	// Continually advertise agent status into message queue
	for {

		// Generate list of locally installed collectors and set FactCollectors
		me.FactCollectors = facts.GetFactCollectors(cfg)

		// Retrieve facts for this agent
		me.Facts = facts.GetFacts(cfg)

		// Advertise this agent
		tc.CommsPackage.AdvertiseAgent(me)

		time.Sleep(15 * time.Second) // TODO(moswalt): make configurable
	}

}
