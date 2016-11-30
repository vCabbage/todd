/*
    ToDD Database Functions

    This file holds the infrastructure for database abstractions in ToDD.

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package db

import (
	"errors"
	"sync"

	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/server/objects"
)

// Errors
var (
	ErrInvalidDBPlugin = errors.New("Invalid DB plugin in config file")
	ErrNotExist        = errors.New("Value does not exist")
)

// Database represents all of the behavior that a ToDD database plugin must support
type Database interface {
	// Agents
	SetAgent(defs.AgentAdvert) error
	GetAgent(uuid string) (*defs.AgentAdvert, error)
	GetAgents() ([]defs.AgentAdvert, error)
	RemoveAgent(uuid string) error

	// Objects
	SetObject(objects.ToddObject) error
	GetObjects(objType string) ([]objects.ToddObject, error)
	DeleteObject(label, objType string) error

	// Group Map
	SetGroupMap(map[string]string) error
	GetGroupMap() (map[string]string, error)

	// Tests
	InitTestRun(testUUID string, testAgentMap map[string]map[string]string) error
	SetAgentTestStatus(testUUID, agentUUID, status string) error
	SetAgentTestData(testUUID, agentUUID string, testData map[string]map[string]interface{}) error
	GetTestStatus(testUUID string) (map[string]string, error)
	GetTestData(testUUID string) (map[string]map[string]map[string]interface{}, error)
}

// New returns an initialized Database.
func New(cfg *config.Config) (Database, error) {
	construct, ok := packages[cfg.DB.Plugin]
	if !ok {
		return nil, ErrInvalidDBPlugin
	}

	return construct(cfg)
}

type constructor func(*config.Config) (Database, error)

var (
	packagesMu sync.Mutex
	packages   = make(map[string]constructor)
)

func register(name string, c constructor) {
	packagesMu.Lock()
	packages[name] = c
	packagesMu.Unlock()
}
