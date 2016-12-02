/*
    ToDD TSDB Functions

    This file holds the infrastructure for TSDB abstractions in ToDD.

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tsdb

import (
	"sync"

	"github.com/pkg/errors"
	"github.com/toddproject/todd/config"
)

// TSDB represents all of the behavior that a ToDD TSDB plugin must support
type TSDB interface {
	WriteData(testUUID, testRunName, groupName string, testData map[string]map[string]map[string]interface{}) error
	Close() error
}

// New will create a new instance of toddTSDB, and load the desired
// databasePackage-compatible comms package into it.
func New(cfg *config.Config) (TSDB, error) { // TODO: return Package instead of *struct embedding Package
	// Load the appropriate TSDB package based on config file
	construct, ok := packages[cfg.TSDB.Plugin]
	if !ok {
		return nil, errors.Errorf("Invalid TSDB plugin in config file: %q", cfg.TSDB.Plugin)
	}

	return construct(cfg)
}

type constructor func(*config.Config) (TSDB, error)

var (
	packagesMu sync.Mutex
	packages   = make(map[string]constructor)
)

func register(name string, c constructor) {
	packagesMu.Lock()
	packages[name] = c
	packagesMu.Unlock()
}
