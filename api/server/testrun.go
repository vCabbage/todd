/*
	ToDD API - manages testruns

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd/api"
	"github.com/toddproject/todd/db"
	"github.com/toddproject/todd/server/objects"
)

// Run will activate an existing testrun
func (s *ServerAPI) Run(w http.ResponseWriter, r *http.Request) {
	// Read the content into a byte array
	// (we're doing this so we can access the JSON contents more than once)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, err)
		return
	}

	// Marshal API data into our struct
	var testRunInfo api.TestRunInfo
	err = json.Unmarshal(body, &testRunInfo)
	if err != nil {
		writeError(w, err)
		return
	}

	// Retrieve list of existing testrun objects
	objectList, err := s.tdb.GetObjects("testrun")
	if err != nil {
		writeError(w, err)
		return
	}

	// See if the requested object name exists within the current object store
	testRunExists := false
	var finalObj objects.ToddObject
	for _, obj := range objectList {
		if obj.GetLabel() == testRunInfo.Name {
			testRunExists = true
			finalObj = obj
			break
		}
	}

	// If testrun object doesn't exist, send error message back to client. Otherwise, proceed with testrun.
	if !testRunExists {
		log.Warnf("Client requested run of testrun object, but %s was not found.", testRunInfo.Name)
		http.NotFound(w, r)
		return
	}

	// Send back the testrun UUID
	testUUID, err := s.Runner.Start(finalObj.(objects.TestRunObject), testRunInfo.SourceOverrides)
	if err != nil {
		writeError(w, err)
		return
	}

	fmt.Fprint(w, testUUID)
}

// TestData will retrieve clean test data by test UUID
func (s *ServerAPI) TestData(w http.ResponseWriter, r *http.Request) {
	// Make sure UUID string is provided
	testUUID := r.URL.Query().Get("testUuid")

	// Make sure UUID string is provided
	if testUUID == "" {
		http.Error(w, "Error, test UUID not provided.", 400)
		return
	}

	testData, err := s.tdb.GetTestData(testUUID)
	if err != nil {
		if err == db.ErrNotExist || len(testData) == 0 {
			http.Error(w, "Error, test UUID not found.", 404)
			return
		}
		writeError(w, err)
		return
	}

	writeJSON(w, testData)
}

// TestStatus streams the agent statuses for the requested test UUID.
func (s *ServerAPI) TestStatus(w http.ResponseWriter, r *http.Request) {
	testUUID := strings.TrimPrefix(r.URL.Path, "/v1/testrun/status/")

	if testUUID == "" {
		http.Error(w, "must provide UUID", http.StatusBadRequest)
	}

	cn, ok := w.(http.CloseNotifier)
	if !ok {
		http.Error(w, "unsupported client", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "unsupported client", http.StatusBadRequest)
		return
	}

	enc := json.NewEncoder(w)

	// Constantly poll for test status, and send statuses to client
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-cn.CloseNotify():
			return
		case <-ticker.C:
			testStatuses, err := s.tdb.GetTestStatus(testUUID)
			if err != nil {
				log.Error("Error retrieving test status:", err)
				return
			}

			err = enc.Encode(testStatuses)
			if err != nil {
				log.Error("Failed to marshal agent test status message:", err)
				return
			}
			flusher.Flush()

			// TODO(mierdin): Need to add failure notification (status of "fail" for any one agent)
		}
	}
}
