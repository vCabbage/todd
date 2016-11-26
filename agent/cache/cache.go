/*
    ToDD Agent Cache

    This package manages the local agent cache. This is a handy
    way of storing some locally significant values on the agent.
    As an example, testrun data is stored here so that the agent
    is able to run tests and cache the post-test metrics in this
    cache until the server acknowledges that it received this data.

    It's also useful for storing key/value type data, such as what
    group the agent is in, it's agent UUID, etc. The Init() function
    will delete and recreate this cache, so no data in this cache
    will persist between restarts of the agent.

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package cache

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3" // This look strange but is necessary - the sqlite package is used indirectly by database/sql
	"github.com/pkg/errors"

	"github.com/toddproject/todd/config"
)

// New returns a new AgentCache.
func New(cfg config.Config) *AgentCache {
	return &AgentCache{dbLoc: filepath.Join(cfg.LocalResources.OptDir, "agent_cache.db")}
}

// AgentCache provides methods for interacting with the on disk cache.
type AgentCache struct {
	// Need similar abstractions to what you did in the tasks package
	dbLoc string
	db    *sql.DB
}

// Init will set up the sqlite database to serve as this agent cache.
func (ac *AgentCache) Init() error {
	// Clean up any old cache data
	err := os.Remove(ac.dbLoc)
	if err != nil && !os.IsNotExist(err) {
		return errors.Wrap(err, "removing existing DB file")
	}

	// Open connection
	db, err := sql.Open("sqlite3", ac.dbLoc)
	if err != nil {
		return errors.Wrap(err, "opening DB file")
	}
	ac.db = db

	// Initialize database
	const sqlStmt = `
		CREATE TABLE testruns (id INTEGER NOT NULL PRIMARY KEY, uuid TEXT, testlet TEXT, args TEXT, targets TEXT, results TEXT);
		CREATE TABLE keyvalue (id INTEGER NOT NULL PRIMARY KEY, key TEXT, value TEXT);
	`

	_, err = db.Exec(sqlStmt)
	if err != nil {
		return fmt.Errorf("%q: %s", err, sqlStmt)
	}
	return nil
}

// Close closes the underlying database connection.
func (ac *AgentCache) Close() error {
	return ac.db.Close()
}
