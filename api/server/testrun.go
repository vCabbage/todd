/*
	ToDD API - manages testruns

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/Mierdin/todd/db"
	"github.com/Mierdin/todd/server/objects"
	"github.com/Mierdin/todd/server/testrun"
)

// Run will activate an existing testrun
func (tapi ToDDApi) Run(w http.ResponseWriter, r *http.Request) {

	// anonymous struct to hold our testRun info
	testRunInfo := struct {
		TestRunName string `json:"testRunName"`
		SourceGroup string `json:"sourceGroup"`
		SourceApp   string `json:"sourceApp"`
		SourceArgs  string `json:"sourceArgs"`
	}{}

	// Marshal API data into our struct
	err = json.NewDecoder(r.Body).(&testRunInfo)
	if err != nil {
		http.Error(w, "Internal Error", 500)
		return
	}

	// Retrieve list of existing testrun objects
	objectList, err := tapi.tdb.GetObjects("testrun")
	if err != nil {
		http.Error(w, "Internal Error", 500)
		return
	}

	// See if the requested object name exists within the current object store
	testRunExists := false
	var finalObj objects.ToddObject
	for i:=0; i<len(objectList); i++ {
		if objectList[i].GetLabel() == testRunInfo.TestRunName {
			testRunExists = true
			finalObj = objectList[i]
			break
		}
	}

	// If testrun object doesn't exist, send error message back to client. Otherwise, proceed with testrun.
	if !testRunExists {
		log.Warnf("Client requested run of testrun object, but %s was not found.", testRunInfo.TestRunName)
		http.NotFound(w, r)
	}

	// Populate sourceOverrideMap dict
	sourceOverrideMap := map[string]string{
		"SourceGroup": testRunInfo.SourceGroup,
		"SourceApp":   testRunInfo.SourceApp,
		"SourceArgs":  testRunInfo.SourceArgs,
	}

	// Send back the testrun UUID
	testUUID := testrun.Start(tapi.cfg, finalObj.(objects.TestRunObject), sourceOverrideMap)
	fmt.Fprint(w, testUUID)
}

// TestData will retrieve clean test data by test UUID
func (tapi ToDDApi) TestData(w http.ResponseWriter, r *http.Request) {
	var testData string
	
	testUUID := r.URL.Query().Get("testUuid")

	// Make sure UUID string is provided
	if testUUID != "" {
		http.Error(w, "Error, test UUID not provided.", 400)
	}

	fmt.Fprint(w, "Error, test UUID not found.")

	testData, err = tapi.tdb.GetCleanTestData(testUUID)
	if err != nil {
		// TODO(kale): Check for key not existing and send 404
		http.Error(w, "Internal Error", 500)
		return
	}

	fmt.Fprint(w, testData)
}
