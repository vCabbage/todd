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

	"github.com/Mierdin/todd/server/objects"
	"github.com/Mierdin/todd/server/testrun"
)

// Run will activate an existing testrun
func (tapi ToDDApi) Run(w http.ResponseWriter, r *http.Request) {

	// Defer the closing of the body
	defer r.Body.Close()

	// Read the content into a byte array
	// (we're doing this so we can access the JSON contents more than once)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	// anonymous struct to hold our testRun info
	testRunInfo := struct {
		TestRunName string `json:"testRunName"`
		SourceGroup string `json:"sourceGroup"`
		SourceApp   string `json:"sourceApp"`
		SourceArgs  string `json:"sourceArgs"`
	}{}

	// Marshal API data into our struct
	err = json.Unmarshal(body, &testRunInfo)
	if err != nil {
		panic(err)
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
	for i := range objectList {
		if objectList[i].GetLabel() == testRunInfo.TestRunName {
			testRunExists = true
			finalObj = objectList[i]
			break
		}
	}

	// If testrun object doesn't exist, send error message back to client. Otherwise, proceed with testrun.
	if !testRunExists {
		log.Warnf("Client requested run of testrun object, but %s was not found.", testRunInfo.TestRunName)
		fmt.Fprint(w, "notfound")
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
	// Make sure UUID string is provided
	uuid := r.URL.Query().Get("testUuid")
	if uuid == "" {
		fmt.Fprint(w, "Error, test UUID not provided.")
		return
	}

	testData, err := tapi.tdb.GetCleanTestData(uuid)
	if err != nil {
		fmt.Fprint(w, "Error, test UUID not found.")
		return
	}

	fmt.Fprint(w, testData)
}
