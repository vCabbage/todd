/*
	ToDD task - test run

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"sync"
	"syscall"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd/agent/cache"
	"github.com/toddproject/todd/agent/testing"
	"github.com/toddproject/todd/config"
)

// ExecuteTestRunTask defines this particular task.
type ExecuteTestRunTask struct {
	BaseTask
	Config    config.Config `json:"-"`
	TestUUID  string        `json:"testuuid"`
	TimeLimit int           `json:"timelimit"`
}

// Run contains the logic necessary to perform this task on the agent. This particular task will execute a
// testrun that has already been installed into the local agent cache. In this context (single agent),
// a testrun will be executed once per target, all in parallel.
func (ett ExecuteTestRunTask) Run() error {

	// gatheredData represents test data from this agent for all targets.
	// Key is target name, value is JSON output from testlet for that target
	// This is reset to a blank map every time ExecuteTestRunTask is called
	gatheredData := map[string]*json.RawMessage{}

	// Use a wait group to ensure that all of the testlets have a chance to finish
	var wg sync.WaitGroup

	// Waiting three seconds to ensure all the agents have their tasks before we potentially hammer the network
	//
	// TODO(mierdin): This is a temporary measure - in the future, testruns will be executed via time schedule,
	// making not only this sleep, but also the entire task unnecessary. Testruns will simply be installed, and
	// executed when the time is right. This is, in part tracked by https://github.com/toddproject/todd/issues/89
	time.Sleep(3000 * time.Millisecond)

	// Retrieve test from cache by UUID
	var ac = cache.NewAgentCache(ett.Config)
	tr, err := ac.GetTestRun(ett.TestUUID)
	if err != nil {
		log.Error(err)
		return errors.New("Problem retrieving testrun from agent cache")
	}

	log.Debugf("IMMA FIRIN MAH LAZER (for test %s) ", ett.TestUUID)

	// Specify size of wait group equal to number of targets
	wg.Add(len(tr.Targets))

	testletPath, err := testing.GetTestletPath(tr.Testlet, ett.Config.LocalResources.OptDir)
	if err != nil {
		return err
	}

	// Execute testlets against all targets asynchronously
	for i := range tr.Targets {

		thisTarget := tr.Targets[i]

		go func() {

			defer wg.Done()

			log.Debugf("Full testlet command and args: '%s %s %s'", testletPath, thisTarget, tr.Args)
			cmd := exec.Command(testletPath, thisTarget, tr.Args)
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

			// Stdout buffer
			cmdOutput := &bytes.Buffer{}
			// Attach buffer to command
			cmd.Stdout = cmdOutput

			// Execute testlet
			cmd.Start()

			done := make(chan error)
			go func() {
				done <- cmd.Wait()
			}()

			// This select statement will block until one of these two conditions are met:
			// - The testlet finishes, in which case the channel "done" will be receive a value
			// - The configured time limit is exceeded (expected for testlets running in server mode)
			select {
			case <-time.After(time.Duration(ett.TimeLimit) * time.Second):

				if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
					log.Errorf("Failed to kill %s after timeout: %s", testletPath, err)
				} else {
					log.Debug("Successfully killed ", testletPath)
				}

				return

			case err := <-done:
				if err != nil {
					log.Errorf("Testlet %s completed with error '%s'", testletPath, err)

					// TODO(mierdin): Handling testrun errors is on my plate as it is, and so
					// this should get addressed properly in the future. The current approach
					// of adding "error" to this map does nothing, as the status is tracked elsewhere
					// gatheredData[thisTarget] = "error"

					return
				}
				log.Debugf("Testlet %s completed without error", testletPath)

			}

			// Record test data
			b := json.RawMessage(cmdOutput.Bytes())
			gatheredData[thisTarget] = &b
		}()
	}

	wg.Wait()

	testdataJSON, err := json.Marshal(gatheredData)
	if err != nil {
		log.Errorf("Failed to marshal post-test data %v", err)
		// TODO(mierdin): This gets triggered for all iperf servers, because they don't provide metrics.
		// Need to figure out a more elegant way of handling this (instead of showing this message)
	}

	// Write test data to agent cache
	err = ac.UpdateTestRunData(ett.TestUUID, string(testdataJSON))
	if err != nil {
		log.Fatal("Failed to install post-test data into cache")
		os.Exit(1)
	} else {
		log.Debugf("Wrote combined post-test data for %s to cache", ett.TestUUID)
	}

	return nil
}
