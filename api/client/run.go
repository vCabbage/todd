/*
    ToDD Client API Calls for "todd run"

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/pkg/errors"
	"github.com/toddproject/todd/api"
	"github.com/toddproject/todd/server/objects"
)

// Run is responsible for activating an existing testrun object
func (c *ClientAPI) Run(sourceOverrides objects.SourceOverrides, testrunName string, displayReport, skipConfirm bool) error {
	// If no subarg was provided, do nothing
	if testrunName == "" {
		return errors.New("Please provide testrun object name to run.")
	}

	if !skipConfirm {
		fmt.Printf("Activate testrun %q? (y/n):", testrunName)
		var userResponse string
		_, err := fmt.Scanln(&userResponse)
		if err != nil {
			return err
		}
		if userResponse != "y" {
			fmt.Println("Aborted.")
			return nil
		}
	}

	testRunInfo := api.TestRunInfo{
		Name:            testrunName,
		SourceOverrides: sourceOverrides,
	}

	// Marshal the final object into JSON
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(testRunInfo)
	if err != nil {
		return err
	}

	// Construct API request, and send POST to server for this object
	url := c.baseURL + "/testrun/run"
	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.Errorf("%s: %s", resp.Status, body)
	}

	testUUID := string(body)

	fmt.Print("\nRUNNING TEST: ", testUUID)
	fmt.Print("\n\n")
	fmt.Println("(Please be patient while the test finishes...)")

	err = c.listenForTestStatus(testUUID)
	if err != nil {
		fmt.Printf("Problem subscribing to testrun updates stream: %s\n", err)
		fmt.Println("Will now watch the testrun metrics API for 45 seconds to see if we get a result that way. Please wait...")
	}

	// Poll for results
	timeout := time.After(45 * time.Second)
	data, err := getRunResult(c.baseURL, testUUID)
	for err != nil {
		select {
		case <-timeout:
			return errors.New("Failed to retrieve test data after 45 seconds. Something must be wrong - quitting.")
		default:
			time.Sleep(1 * time.Second)
			data, err = getRunResult(c.baseURL, testUUID)
		}
	}

	fmt.Printf("\n\nDone.\n")

	// display it to the user if desired
	if sourceOverrides.Group != "" || displayReport {
		var buf bytes.Buffer
		err := json.Indent(&buf, data, "", "  ")
		if err != nil {
			fmt.Printf("error %q: %v\n", string(data), err)
		}
		buf.WriteTo(os.Stdout)
		fmt.Println()
	}

	return nil
}

// getRunResult collects the results of test run from the server's REST API
func getRunResult(baseURL string, testUUID string) ([]byte, error) {
	// Go back and get our testrun data
	url := fmt.Sprintf("%s/testdata?testUuid=%s", baseURL, testUUID)

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	// Defer the closing of the body
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	return ioutil.ReadAll(resp.Body)
}

// listenForTestStatus connects to the server's test event stream and prints the progression
//
// This blocks until all agents have finished or an error occurs.
func (c *ClientAPI) listenForTestStatus(uuid string) error {
	resp, err := c.http.Get(c.baseURL + "/testrun/status/" + uuid)
	if err != nil {
		return err
	}
	defer io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	dec := json.NewDecoder(resp.Body)

	for recordCount := -1; dec.More(); {
		var statuses map[string]string
		err := dec.Decode(&statuses)
		if err != nil {
			return err
		}

		if recordCount < 1 {
			recordCount = len(statuses)
		}

		init, ready, testing, finished := 0, 0, 0, 0
		for _, status := range statuses {
			switch status {
			case "init":
				init++
			case "ready":
				ready++
			case "testing":
				testing++
			case "finished":
				finished++
			default:
				return errors.New("Invalid status received.")
			}
		}

		// Print the status line (note the \r which keeps the same line in place on the terminal)
		fmt.Printf(
			"\r %[1]s INIT: (%[3]d/%[2]d)  READY: (%[4]d/%[2]d)  TESTING: (%[5]d/%[2]d)  FINISHED: (%[6]d/%[2]d)",
			time.Now(),
			recordCount,
			init,
			ready,
			testing,
			finished,
		)

		if finished == recordCount {
			break
		}
	}

	return nil
}
