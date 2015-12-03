/*
   ToDD Database Functions

    This file holds the infrastructure for database abstractions in ToDD.

   Copyright 2015 - Matt Oswalt
*/

package db

import (
    "github.com/mierdin/todd/agent/defs"
    "github.com/mierdin/todd/config"
)

// DatabasePackage will ensure that whatever specific DB struct is loaded at compile time
// will support all of the abstract database functions that we need to support.
type DatabasePackage interface {

    // (no args)
    Init()

    // (agent advertisement to set)
    SetAgent(defs.AgentAdvert)

    // (no args)
    GetAgents(string) []defs.AgentAdvert

    // (agent advertisement to remove)
    RemoveAgent(defs.AgentAdvert)

    SetObject(string, []map[string]string)
    GetObject()
    RemoveObject()
}

// ToddDatabase is a struct to hold anything that satisfies the databasePackage interface
type ToddDatabase struct {
    DatabasePackage
}

// NewToDDComms will create a new instance of ToDDComms, and load the desired
// databasePackage-compatible comms package into it.
func NewToddDB(cfg config.Config) *ToddDatabase {

    // Create ToddDatabase instance
    var tdb ToddDatabase

    // Set DatabasePackage to desired package
    // TODO(mierdin): Need to make this dynamic based on config
    var etcddb EtcdDB
    etcddb.config = cfg
    tdb.DatabasePackage = etcddb

    return &tdb
}
