/*
    ToDD Agent Cache - working with testruns

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package cache

import (
	"encoding/json"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	"github.com/toddproject/todd/agent/defs"
)

// InsertTestRun places a new test run into the agent cache
func (ac *AgentCache) InsertTestRun(tr defs.TestRun) error {
	// Marshal our string slices to be stored in the database
	jsonTargets, err := json.Marshal(tr.Targets)
	if err != nil {
		return errors.Wrap(err, "marshaling testrun targets into JSON")
	}
	jsonArgs, err := json.Marshal(tr.Args)
	if err != nil {
		return errors.Wrap(err, "marshaling testrun args into JSON")
	}

	_, err = ac.db.Exec("INSERT INTO testruns(uuid, testlet, args, targets) VALUES(?, ?, ?, ?)",
		tr.UUID, tr.Testlet, string(jsonArgs), string(jsonTargets))
	if err != nil {
		return errors.Wrap(err, "executing new testrun insert")
	}

	log.Info("Inserted new testrun into agent cache - ", tr.UUID)

	return nil
}

// GetTestRun retrieves a testrun from the agent cache by its UUID
func (ac *AgentCache) GetTestRun(uuid string) (*defs.TestRun, error) {
	rows, err := ac.db.Query("SELECT testlet, args, targets FROM testruns WHERE uuid = ?", uuid)
	if err != nil {
		return nil, errors.Wrap(err, "creating query for selecting testrun")
	}
	defer rows.Close()

	tr := &defs.TestRun{UUID: uuid}
	// TODO(mierdin): This may be unnecessary - rows.Scan() might allow you to pass this in as a byteslice. Experiment with this
	var argsJSON, targetsJSON string
	for rows.Next() {
		err = rows.Scan(&tr.Testlet, &argsJSON, &targetsJSON)
		if err != nil {
			return tr, errors.Wrap(err, "scanning testrun data from database")
		}

		err = json.Unmarshal([]byte(argsJSON), &tr.Args)
		if err != nil {
			return tr, errors.Wrap(err, "unmarshaling testrun args from JSON")
		}

		err = json.Unmarshal([]byte(targetsJSON), &tr.Targets)
		if err != nil {
			return tr, errors.Wrap(err, "unmarshaling testrun targets from JSON")
		}
	}

	log.Infof("Found testrun %q running testlet %q\n", tr.UUID, tr.Testlet)

	return tr, nil
}

// UpdateTestRunData will update an existing testrun entry in the agent cache with the post-test
// metrics dataset that corresponds to that testrun (by testrun UUID)
func (ac *AgentCache) UpdateTestRunData(uuid string, testData string) error {
	_, err := ac.db.Exec("UPDATE testruns SET results = ? WHERE uuid = ?", testData, uuid)
	if err != nil {
		return errors.Wrap(err, "preparing new UpdateTestRunData action")
	}

	log.Infof("Inserted test data for %q into cache\n", uuid)

	return nil
}

// DeleteTestRun will remove an entire testrun entry from teh agent cache by UUID
func (ac *AgentCache) DeleteTestRun(uuid string) error {
	_, err := ac.db.Exec("DELETE FROM testruns WHERE uuid = ?", uuid)
	return errors.Wrap(err, "preparing new DeleteTestRun action")
}

// GetFinishedTestRuns returns a map of test UUIDS (keys) and the corresponding post-test metric data for those UUIDs (values)
// The metric data is stored as a string containing JSON text, so this is what's placed into this map (meaning JSON parsing is
// not performed in this function)
func (ac *AgentCache) GetFinishedTestRuns() (map[string]string, error) {
	rows, err := ac.db.Query(`SELECT uuid, results FROM testruns WHERE results != ""`)
	if err != nil {
		return nil, errors.Wrap(err, "creating query for selecting finished testruns")
	}
	defer rows.Close()

	retmap := make(map[string]string)
	var uuid, testdata string
	for rows.Next() {
		err = rows.Scan(&uuid, &testdata)
		if err != nil {
			return nil, errors.Wrap(err, "scanning testrun data from database")
		}
		log.Debug("Found ripe testrun: ", uuid)
		retmap[uuid] = testdata
	}

	return retmap, nil
}
