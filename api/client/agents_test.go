/*
   Unit testing for ToDD Client API - agents.go

   Copyright 2016 Matt Oswalt. Use or modification of this
   source code is governed by the license provided here:
   https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Mierdin/todd/agent/defs"
)

var testAgent = defs.AgentAdvert{
	Uuid:        "aaaaaaaaaaaaaaaaaaaaaa",
	DefaultAddr: "192.168.0.1",
	Expires:     27000000000,
	LocalTime:   time.Now(),
	Facts: map[string][]string{
		"Hostname": []string{
			"testmachine",
		},
	},
	FactCollectors: map[string]string{
		"get_addresses": "aaaaaaaaaaaaaaaaaaaaaa",
		"get_hostname":  "aaaaaaaaaaaaaaaaaaaaaa",
	},
	Testlets: map[string]string{
		"iperf": "aaaaaaaaaaaaaaaaaaaaaa",
		"ping":  "aaaaaaaaaaaaaaaaaaaaaa",
	},
}

// TestAgents tests the ability for the Agents client API call to function correctly
func TestAgents(t *testing.T) {

	testAgentSlice := []defs.AgentAdvert{testAgent}

	agentJson, err := json.Marshal(testAgentSlice)
	if err != nil {
		t.Error(err)
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(agentJson))
	}))
	defer ts.Close()

	agentsUrl := fmt.Sprintf("%s/v1/agent", ts.URL)

	var capi ClientApi
	err, agents := capi.Agents(
		map[string]string{
			"host": strings.Split(strings.Replace(strings.Replace(agentsUrl, "http://", "", 1), "/v1/agent", "", 1), ":")[0],
			"port": strings.Split(strings.Replace(strings.Replace(agentsUrl, "http://", "", 1), "/v1/agent", "", 1), ":")[1],
		}, "",
	)
	if err != nil {
		t.Error(err)
	}

	if len(agents) != 1 {
		t.Error("Incorrect number of agents found")
	}
}

// agentTests is a "table" of test cases to apply to TestDisplayAgents
var agentTests = []struct {
	arg1 []defs.AgentAdvert
	arg2 bool
}{
	{[]defs.AgentAdvert{testAgent}, false},
	{[]defs.AgentAdvert{testAgent}, true},
	{[]defs.AgentAdvert{}, false},
}

// TestDisplayAgents iterates over the test cases and runs DisplayAgents on each
func TestDisplayAgents(t *testing.T) {
	var capi ClientApi
	for _, test := range agentTests {
		err := capi.DisplayAgents(test.arg1, test.arg2)
		if err != nil {
			t.Fatal(err)
		}
	}
}
