/*
    ToDD Test Runs

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package testrun

import (
	"context"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/agent/tasks"
	"github.com/toddproject/todd/comms"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/db"
	"github.com/toddproject/todd/hostresources"
	"github.com/toddproject/todd/server"
	"github.com/toddproject/todd/server/objects"
	"github.com/toddproject/todd/server/tsdb"
)

// TestRunner provides logic for running tests.
type TestRunner struct {
	cfg   *config.Config
	srv   *server.Server
	db    db.Database
	comms comms.Comms
	tsdb  tsdb.TSDB
}

// New returns a new TestRunner.
func New(cfg *config.Config, srv *server.Server, d db.Database, comm comms.Comms, ts tsdb.TSDB) *TestRunner {
	return &TestRunner{
		cfg:   cfg,
		srv:   srv,
		db:    d,
		comms: comm,
		tsdb:  ts,
	}
}

// Start executes a test.
func (t *TestRunner) Start(trObj objects.TestRunObject, sourceOverrides objects.SourceOverrides) (string, error) {
	// Generate UUID for test
	testUUID, err := hostresources.GenerateUUID()
	if err != nil {
		return "", errors.Wrap(err, "generating UUID")
	}

	// Retrieve current group map
	allGroupMap, err := t.db.GetGroupMap()
	if err != nil {
		return "", errors.Wrap(err, "retrieving group map")
	}

	sourceOverrides.Apply(trObj.Spec.Source)

	//testAgentMap is a custom map that only contains agents in the one or two groups relevant to this test
	//The outer map should contain two keys, "targets", and "sources". The values for each key should be another
	// map that uses the agent UUID for keys, and the group for those UUIDs as values.
	testAgentMap := map[string]map[string]string{
		"sources": make(map[string]string),
		"targets": make(map[string]string),
	}

	// Here, we iterate over ALL of the agents, and pick out the ones that are part of this test, as well as what group they're in.
	for agent, group := range allGroupMap {
		// If our target type is group, and the group this agent is in matches the target group provided in the testrun object, add it to our map
		if (trObj.Spec.TargetType == "group" && group == trObj.Spec.Target.Map["name"]) ||
			group == trObj.Spec.Source["name"] {
			testAgentMap["sources"][agent] = group
		}
	}

	// Reject this topology if there aren't the right number of agents registered in this topology.
	if (trObj.Spec.TargetType == "group" && len(testAgentMap["targets"]) == 0) ||
		len(testAgentMap["sources"]) == 0 {
		return "", errors.New("invalid topology")
	}

	// Start listening for responses from agents
	ctx, cancel := context.WithCancel(context.Background())
	msgs, err := t.comms.ListenForResponses(ctx)
	if err != nil {
		cancel()
		return "", err
	}
	go func() {
		for msg := range msgs {
			err := t.srv.HandleAgentResponse(msg)
			if err != nil {
				log.Error("Error handling agent response:", err)
			}
		}
	}()

	// Initialize test in database. This will create an entry for this test under the UUID we just created, and will also write the
	// list of agents participating in this test, with some kind of default status, for other goroutines to update with a further status.
	err = t.db.InitTestRun(testUUID, testAgentMap)
	if err != nil {
		cancel()
		return "", errors.Wrap(err, "initializing testrun in DB")
	}

	// Prepare testrun instruction for our source agents
	sourceTr := defs.TestRun{
		UUID:    testUUID,
		Testlet: trObj.Spec.Source["app"],
		Args:    trObj.Spec.Source["args"],
		Targets: trObj.Spec.Target.Slice, // Default to uncontrolled target, target IPs in testrun object
	}

	// If target type is actually group, get IPs from DB.
	if trObj.Spec.TargetType == "group" {
		agents, err := t.db.GetAgents()
		if err != nil {
			cancel()
			return "", errors.Wrap(err, "retrieving agents")
		}
		for _, agent := range agents {
			if _, ok := testAgentMap["targets"][agent.UUID]; ok {
				sourceTr.Targets = append(sourceTr.Targets, agent.DefaultAddr)
			}
		}
	}

	// Prepare a task for carrying the testrun instruction to the agent
	itrTask := &tasks.InstallTestRun{
		BaseTask: tasks.BaseTask{Type: "InstallTestRun"},
		TR:       sourceTr,
	}

	// Send testrun to each agent UUID in the sources group
	// TODO(mierdin): this is something I'd like to improve in the future. Right now this works, and is sort-of resilient, since
	// the testrun will require a response from each agent before actually moving on with execution, but I'd like something better.
	// Something that feels more like a true distributed system. Perfect is the enemy of good, however, and this works well for a prototype.
	for uuid := range testAgentMap["sources"] {
		err = t.comms.SendTask(uuid, itrTask)
		if err != nil {
			cancel()
			return "", errors.Wrapf(err, "sending task to %q", uuid)
		}
	}

	if trObj.Spec.TargetType == "group" {
		// If this testrun is targeted at another todd group, we want to send testrun tasks to those as well
		targetTr := defs.TestRun{
			UUID:    testUUID,
			Targets: []string{"0.0.0.0"}, // Targets are typically running some kind of ongoing service, so we send a single target of 0.0.0.0 to indicate this.
			Testlet: trObj.Spec.Target.Map["app"],
			Args:    trObj.Spec.Target.Map["args"],
		}

		itrTask := &tasks.InstallTestRun{
			BaseTask: tasks.BaseTask{Type: "InstallTestRun"},
			TR:       targetTr,
		}

		// Send testrun to each agent UUID in the targets group
		for uuid := range testAgentMap["targets"] {
			t.comms.SendTask(uuid, itrTask)
			if err != nil {
				cancel()
				return "", errors.Wrapf(err, "sending task to %q", uuid)
			}
		}
	}

	go t.executeTestRun(ctx, testAgentMap, testUUID, trObj, cancel, !sourceOverrides.AnySet())

	// Return the testUuid so that the client can subscribe to it.
	return testUUID, nil
}

// executeTestRun will perform three things:
//
// - Monitor the database to determine which agents have which statuses
// - When the status for all agents is "ready", it will send execution tasks to one or both groups
// - It will continue to monitor, and when all agents have finished, it will pull the "leash" to stop the TCP stream to the client
// - After pulling the leash, it will call the function that will aggregate the test data and upload to a third party service
func (t *TestRunner) executeTestRun(ctx context.Context, testAgentMap map[string]map[string]string, testUUID string, trObj objects.TestRunObject, cancel func(), updateTSDB bool) {
	// First, let's just keep retrieving statuses until all of the agents are reporting ready
	err := t.agentsInStatus(ctx, testUUID, "ready")
	if err != nil {
		cancel()
		log.Error(err)
		return
	}

	// If this is a group target type, we want to make sure that the targets are set up and reporting a status of "testing"
	// before we spin up the source tests
	if trObj.Spec.TargetType == "group" {
		// Next, we want to wait to make sure that the targets are all "testing" before instructing the source group to execute
		err = t.targetsReady(ctx, testUUID, testAgentMap, trObj)
		if err != nil {
			cancel()
			log.Error(err)
			return
		}
	}

	// The targets are ready; execute testing on the source agents
	sourceTask := &tasks.ExecuteTestRun{
		BaseTask:  tasks.BaseTask{Type: "ExecuteTestRun"},
		TestUUID:  testUUID,
		TimeLimit: 30,
	}

	// Send testrun to each agent UUID in the targets group
	for uuid := range testAgentMap["sources"] {
		err = t.comms.SendTask(uuid, sourceTask)
		if err != nil {
			cancel()
			log.Errorf("Error sending execute task to %q: %v", uuid, err)
			return
		}
	}

	if trObj.Spec.TargetType == "group" {
		targetTask := &tasks.ExecuteTestRun{
			BaseTask:  tasks.BaseTask{Type: "ExecuteTestRun"},
			TestUUID:  testUUID,
			TimeLimit: t.cfg.Testing.Timeout,
		}

		// Send testrun to each agent UUID in the targets group
		for uuid := range testAgentMap["targets"] {
			t.comms.SendTask(uuid, targetTask)
			if err != nil {
				cancel()
				log.Errorf("Error sending execute task to target %q: %v", uuid, err)
				return
			}
		}
	}

	// Let's wait once more until all agents are stored in the database with a status of "finished"
	err = t.agentsInStatus(ctx, testUUID, "finished")
	if err != nil {
		cancel()
		log.Error(err)
		return
	}

	if !updateTSDB {
		return
	}

	testData, err := t.db.GetTestData(testUUID)
	if err != nil {
		log.Error("Error retrieving agent test data:", err)
		return
	}

	err = t.tsdb.WriteData(testUUID, trObj.Label, trObj.Spec.Source["name"], testData)
	if err != nil {
		log.Error("Problem writing metrics to TSDB:", err)
	}
}

func (t *TestRunner) agentsInStatus(ctx context.Context, testUUID, status string) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return errors.New("Test Run Cancelled")
		case <-ticker.C:
			// TODO: Update when new status from client
			testStatuses, err := t.db.GetTestStatus(testUUID)
			if err != nil {
				if err == db.ErrNotExist {
					continue
				}
				log.Error("Error retrieving test status:", err)
			}

			for _, status := range testStatuses {
				switch status {
				case "fail":
					return errors.New("Agent reported failure during testing")
				case status:
					continue
				}
			}
			return nil
		}
	}
}

func (t *TestRunner) targetsReady(ctx context.Context, testUUID string, testAgentMap map[string]map[string]string, trObj objects.TestRunObject) error {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return errors.New("Test Run Cancelled")
		case <-ticker.C:
			testStatuses, err := t.db.GetTestStatus(testUUID)
			if err != nil {
				if err == db.ErrNotExist {
					continue
				}
				log.Fatalf("Error retrieving test status: %v", err)
			}

			// We're creating this map to hold ONLY the targets that are in testStatuses (which also contains sources)
			targetStatuses := make(map[string]string)

			// Iterate all of the sources and agents
			for agent, status := range testStatuses {
				// if this agent is present in the group map
				if val, ok := testAgentMap["targets"][agent]; ok && val == trObj.Spec.Target.Map["name"] {
					targetStatuses[agent] = status
				}
			}

			for _, status := range testStatuses {
				switch status {
				case "fail":
					return errors.New("Agent reported failure during testing")
				case "testing":
					continue
				}
			}
			return nil
		}
	}
}
