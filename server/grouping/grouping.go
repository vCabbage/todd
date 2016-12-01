/*
    ToDD agent grouping mechanism

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package grouping

import (
	"net"
	"regexp"

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"

	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/agent/tasks"
	"github.com/toddproject/todd/comms"
	"github.com/toddproject/todd/config"
	"github.com/toddproject/todd/db"
	"github.com/toddproject/todd/server/objects"
)

// CalculateGroups is a function designed to ingest a list of group objects, a collection of agents, and
// return a map that contains the resulting group for each agent UUID.
func CalculateGroups(cfg *config.Config, tdb db.Database, tc comms.Comms) error { // TODO: Currently an agent can only be in a single group, allow multiple
	// Retrieve all currently active agents
	agents, err := tdb.GetAgents()
	if err != nil {
		return errors.Wrap(err, "retrieving agents")
	}

	// Retrieve all objects with type "group"
	groupObjs, err := tdb.GetObjects("group")
	if err != nil {
		return errors.Wrap(err, "retrieving groups")
	}

	// Cast retrieved slice of ToddObject interfaces to actual GroupObjects
	groups := make([]objects.GroupObject, len(groupObjs))
	for i, gobj := range groupObjs {
		groups[i] = gobj.(objects.GroupObject)
	}

	// groupmap contains the uuid-to-groupname mappings to be used for test runs
	groupmap := make(map[string]string)

	// This slice will hold all of the agents that are sad because they didn't get into a group
	var lonelyAgents []defs.AgentAdvert

next:
	for _, agent := range agents {
		for _, group := range groups {
			// See if this agent is in this group
			ok, err := isInGroup(group.Spec.Matches, agent.Facts)
			if err != nil {
				return errors.Wrapf(err, "checking group for %q", agent.UUID)
			}
			if !ok {
				continue
			}

			// Insert this group name ("Label") into groupmap under the key of the UUID for the agent that belongs to it
			log.Debugf("Agent %s is in group %s", agent.UUID, group.Label)
			groupmap[agent.UUID] = group.Label
			continue next
		}

		// The "continue next" should prohibit all agents that have a group from getting to this point,
		// so the only ones left do not have a group.
		lonelyAgents = append(lonelyAgents, agent)
	}

	// Write results to database
	err = tdb.SetGroupMap(groupmap)
	if err != nil {
		log.Fatalf("Error setting group map: %v", err)
	}

	// Send notifications to each agent to let them know what group they're in, so they can cache it
	for uuid, groupName := range groupmap {
		setGroupTask := &tasks.SetGroup{
			BaseTask:  tasks.BaseTask{Type: "SetGroup"},
			GroupName: groupName,
		}
		tc.SendTask(uuid, setGroupTask)
	}

	// need to send a message to all agents that weren't in groupmap to set their group to nothing
	for _, agent := range lonelyAgents {
		setGroupTask := &tasks.SetGroup{
			BaseTask:  tasks.BaseTask{Type: "SetGroup"},
			GroupName: "",
		}
		tc.SendTask(agent.UUID, setGroupTask)
	}

	return nil
}

type matcher func(value string, facts map[string][]string) (bool, error)

var matchers = map[string]matcher{
	"hostname": func(regex string, facts map[string][]string) (bool, error) {
		// Continue if no "Hostname" fact present
		hostnames, ok := facts["Hostname"]
		if !ok {
			return false, nil
		}

		exp, err := regexp.Compile(regex)
		if err != nil {
			return false, errors.Wrapf(err, "parsing %s", regex)
		}

		for _, hostname := range hostnames {
			if exp.MatchString(hostname) {
				return true, nil
			}
		}

		return false, nil
	},
	"within_subnet": func(subnet string, facts map[string][]string) (bool, error) {
		// Continue if no "Addresses" fact present
		addresses, ok := facts["Addresses"]
		if !ok {
			return false, nil
		}

		// First, we retrieve the desired subnet from the grouping object, and convert to net.IPNet
		_, desiredNet, err := net.ParseCIDR(subnet)
		if err != nil {
			return false, errors.Wrapf(err, "parsing %s", subnet)
		}

		for _, address := range addresses {
			if desiredNet.Contains(net.ParseIP(address)) {
				return true, nil
			}
		}

		return false, nil
	},
}

// isInGroup takes a set of match statements (typically present in a group object definition) and a map of a single agent's facts,
// and returns True if one of the match statements validated against this map of facts. In short, this function can tell you if an agent
// is in a given group. This means that ToDD stops at the first successful match.
//
// This function is currently written to statically provide two mechanisms for matching:
//
// - hostname
// - within_subnet
//
// In the future, this functionality will be made much more generic, in order to take advantage of any user-defined fact collectors.
func isInGroup(matchStatements []map[string]string, facts map[string][]string) (bool, error) {
	match := false

	// Iterate over the match statements
	for _, matchStatement := range matchStatements {
		for name, value := range matchStatement {
			matcher, ok := matchers[name]
			if !ok {
				return false, errors.Errorf("unknown matcher: %q", name)
			}

			ok, err := matcher(value, facts)
			if err != nil {
				return false, err
			}
			match = match || ok
		}
	}

	return match, nil
}
