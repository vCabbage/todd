/*
    ToDD Agent Cache - working with keyvalues

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package cache

import (
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

// GetKeyValue will retrieve a value from the agent cache using a key string
func (ac *AgentCache) GetKeyValue(key string) (string, error) {
	log.Debugf("Retrieving value of key - %s\n", key)

	// First, see if the key exists.
	rows, err := ac.db.Query("SELECT value FROM keyvalue WHERE key = ?", key)
	if err != nil {
		return "", errors.Wrap(err, "querying DB")
	}
	defer rows.Close()

	var value string
	for rows.Next() {
		err = rows.Scan(&value)
		if err != nil {
			return "", errors.Wrap(err, "scanning values retrieved from DB")
		}
	}
	return value, nil
}

// SetKeyValue sets a KeyValue pair within the agent cache
func (ac *AgentCache) SetKeyValue(key, value string) error {
	log.Debugf("Writing keyvalue pair to agent cache - %s:%s\n", key, value)

	// First, see if the key exists.
	rows, err := ac.db.Query("SELECT count(1) FROM keyvalue WHERE KEY = ?", key)
	if err != nil {
		return errors.Wrap(err, "querying count from DB")
	}
	defer rows.Close()

	var rowCount int
	for rows.Next() {
		err = rows.Scan(&rowCount)
		if err != nil {
			return errors.Wrap(err, "scanning rowCount from DB")
		}
	}

	if rowCount != 1 {
		tx, err := ac.db.Begin()
		if err != nil {
			return errors.Wrap(err, "starting DB transaction")
		}

		if rowCount > 1 {
			// If there is MORE than one row, we should delete the extras first
			// TODO(mierdin): Is this really necessary?
			log.Warn("Extra keyvalue pair detected. Deleting and inserting new record.")

			_, err = tx.Exec("DELETE FROM keyvalue WHERE KEY = ?", key)
			if err != nil {
				tx.Rollback()
				return errors.Wrap(err, "deleteing keyvalues")
			}
		}

		_, err = tx.Exec("INSERT INTO keyvalue(key, value) values(?, ?)", key, value)
		if err != nil {
			tx.Rollback()
			return errors.Wrap(err, "inserting keyvalue into DB")
		}

		return errors.Wrap(tx.Commit(), "commmitting transaction")
	}

	_, err = ac.db.Exec("UPDATE keyvalue SET value = ? WHERE key = ?", value, key)
	return errors.Wrap(err, "updating keyvalue")
}
