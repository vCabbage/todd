/*
    ToDD Client API Calls for "todd run"

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Run is responsible for activating an existing testrun object
func (capi ClientApi) Run(conf map[string]string, testrunName string, displayReport, skipConfirm bool) {

	sourceGroup := conf["sourceGroup"]
	sourceApp := conf["sourceApp"]
	sourceArgs := conf["sourceArgs"]

	// If no subarg was provided, do nothing
	if testrunName == "" {
		fmt.Println("Please provide testrun object name to run.")
		os.Exit(1)
	}

	if !skipConfirm {
		fmt.Printf("Activate testrun \"%s\"? (y/n):", testrunName)
		var userResponse string
		_, err := fmt.Scanln(&userResponse)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if userResponse != "y" {
			fmt.Println("Aborted.")
			os.Exit(0)
		}
	}

	// anonymous struct to hold our testRun info
	testRunInfo := struct {
		TestRunName string `json:"testRunName"`
		SourceGroup string `json:"sourceGroup"`
		SourceApp   string `json:"sourceApp"`
		SourceArgs  string `json:"sourceArgs"`
	}{
		testrunName,
		sourceGroup,
		sourceApp,
		sourceArgs,
	}

	// Marshal the final object into JSON
	json_str, err := json.Marshal(testRunInfo)
	if err != nil {
		panic(err)
	}
	// Construct API request, and send POST to server for this object
	var url string
	url = fmt.Sprintf("http://%s:%s/v1/testrun/run", conf["host"], conf["port"])

	var jsonByte = []byte(json_str)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Print error if not 200 OK
	if resp.Status != "200 OK" {
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		os.Exit(1)
	}

	serverResponse, _ := ioutil.ReadAll(resp.Body)

	switch string(serverResponse) {
	case "notfound":
		fmt.Println("ERROR - Specified testrun object not found.")
		os.Exit(1)

	case "invalidtopology":
		fmt.Println("ERROR - Not enough agents are in the groups specified by the testrun")
		os.Exit(1)

	case "failure":
		fmt.Println("ERROR - some kind of error was encountered on the server. Test was not run.")
		os.Exit(1)

	default:

		fmt.Println("")

		testUuid := string(serverResponse)

		firstMessage := false
		var recordCount int

		fmt.Print("RUNNING TEST: ", testUuid)
		fmt.Print("\n\n")

		fmt.Println("(Please be patient while the test finishes...)\n")

		retries := 0
	retry:
		// connect to this socket
		conn, err := net.Dial("tcp", fmt.Sprintf("%s:8081", conf["host"]))
		if err != nil {

			retries = retries + 1
			time.Sleep(1000 * time.Millisecond)
			if retries > 5 {
				fmt.Println("Failed to subscribe to test event stream after several retries.")
				fmt.Println("Will now watch the testrun metrics API for 45 seconds to see if we get a result that way. Please wait...")
				goto tcpfailed
			} else {
				fmt.Println("Failed to subscribe to test event stream. Retrying...")
				goto retry
			}
		}

		for {

			// listen for reply
			message, err := bufio.NewReader(conn).ReadString('\n')

			// If an error is raised, it's probably because the server killed the connection
			if err != nil {
				// TODO(mierdin): This doesn't really tell us if the connection died because of an error or not
				break
			}

			var statuses map[string]string
			err = json.Unmarshal([]byte(message), &statuses)
			// TODO (mierdin): Handle error

			if firstMessage == false {
				recordCount = len(statuses)
				firstMessage = true
			}

			init, ready, testing, finished := 0, 0, 0, 0
			for _, status := range statuses {

				switch status {
				case "init":
					init += 1
				case "ready":
					ready++
				case "testing":
					testing++
				case "finished":
					finished++
				default:
					fmt.Println("Invalid status recieved.")
					os.Exit(1)
				}
			}

			// Print the status line (note the \r which keeps the same line in place on the terminal)
			fmt.Printf(
				"\r %s INIT: (%s/%s)  READY: (%s/%s)  TESTING: (%s/%s)  FINISHED: (%s/%s)",
				time.Now(),
				strconv.Itoa(init),
				strconv.Itoa(recordCount),
				strconv.Itoa(ready),
				strconv.Itoa(recordCount),
				strconv.Itoa(testing),
				strconv.Itoa(recordCount),
				strconv.Itoa(finished),
				strconv.Itoa(recordCount),
			)

			// Send an ack back to the server to let it know we're alive
			fmt.Fprintf(conn, "ack\n")

		}

		retries = 0

	tcpfailed:

		// Go back and get our testrun data
		url = fmt.Sprintf("http://%s:%s/v1/testdata?testUuid=%s", conf["host"], conf["port"], testUuid)

		// Build the request
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			panic(err)
		}

		// Send the request via a client
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			panic(err)
		}

		// Defer the closing of the body
		defer resp.Body.Close()
		// Read the content into a byte array
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			panic(err)
		}

		var data map[string]map[string]map[string]string
		err = json.Unmarshal([]byte(body), &data)

		b, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			fmt.Println("error:", err)
		}
		if string(b) == "null" {

			// More than likely, testrun has not yet completed. So, let's wait a bit and retry.
			retries = retries + 1
			time.Sleep(1000 * time.Millisecond)
			if retries > 45 {
				fmt.Println("Failed to retrieve test data after 45 seconds. Something must be wrong - quitting.")
				os.Exit(1)
			} else {
				goto tcpfailed
			}

		} else {

			fmt.Println("\n\nDone.\n")

			// display it to the user if desired
			if sourceGroup != "" || displayReport {
				fmt.Println(string(b))
			}
		}

		os.Exit(0)
	}

}
