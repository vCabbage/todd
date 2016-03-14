/*
    ToDD Database Functions

    This file holds the infrastructure for database abstractions in ToDD.

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package db

import (
	"os"

	"github.com/Mierdin/todd/agent/defs"
	"github.com/Mierdin/todd/config"
	"github.com/Mierdin/todd/server/objects"
	log "github.com/Sirupsen/logrus"
)

// DatabasePackage represents all of the behavior that a ToDD database plugin must support
type DatabasePackage interface {

	// (no args)
	Init()

	// (agent advertisement to set)
	SetAgent(defs.AgentAdvert)

	GetAgent(string) defs.AgentAdvert
	GetAgents() []defs.AgentAdvert

	// (agent advertisement to remove)
	RemoveAgent(defs.AgentAdvert)

	SetObject(objects.ToddObject)
	GetObjects(string) []objects.ToddObject
	DeleteObject(string, string)

	GetGroupMap() map[string]string
	SetGroupMap(map[string]string)

	// Testing
	InitTestRun(string, map[string]map[string]string) error
	SetAgentTestStatus(string, string, string) error
	GetTestStatus(string) map[string]string
	SetAgentTestData(string, string, string) error
	GetAgentTestData(string, string) map[string]string
	WriteCleanTestData(string, string)
	GetCleanTestData(string) string
}

// toddDatabase is a struct to hold anything that satisfies the databasePackage interface
type toddDatabase struct {
	DatabasePackage
}

// NewToDDComms will create a new instance of toddDatabase, and load the desired
// databasePackage-compatible comms package into it.
func NewToddDB(cfg config.Config) *toddDatabase {

	// Create toddDatabase instance
	var tdb toddDatabase

	// Load the appropriate DB package based on config file
	switch cfg.DB.Plugin {
	case "etcd":
		tdb.DatabasePackage = newEtcdDB(cfg)
	default:
		log.Error("Invalid DB plugin in config file")
		os.Exit(1)
	}

	return &tdb
}
