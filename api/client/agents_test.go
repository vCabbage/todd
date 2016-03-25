/*
   Unit testing for Agents (Client API)

   Copyright 2016 Matt Oswalt. Use or modification of this
   source code is governed by the license provided here:
   https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Mierdin/todd/agent/defs"
)

// TestAgents tests the ability for the Agents client API call to function correctly
func TestAgents(t *testing.T) {

	agentJSON := `
[
  {
    "Uuid": "aaaaaaaaaaaaaaaaaaaaaa",
    "DefaultAddr": "192.168.0.1",
    "Expires": 27000000000,
    "LocalTime": "2016-03-25T08:03:11.378211992Z",
    "Facts": {
      "Hostname": [
        "testmachine"
      ]
    },
    "FactCollectors": {
      "get_addresses": "aaaaaaaaaaaaaaaaaaaaaa",
      "get_hostname": "aaaaaaaaaaaaaaaaaaaaaa"
    },
    "Testlets": {
      "iperf": "aaaaaaaaaaaaaaaaaaaaaa",
      "ping": "aaaaaaaaaaaaaaaaaaaaaa"
    }
  }
]
`

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, agentJSON)
	}))
	defer ts.Close()

	agentsUrl := fmt.Sprintf("%s/v1/agent", ts.URL)

	var capi ClientApi
	err, agents := capi.Agents(
		map[string]string{
			"host": strings.Split(strings.Replace(agentsUrl, "http://", "", 1), ":")[0],
			"port": strings.Split(strings.Replace(agentsUrl, "http://", "", 1), ":")[1],
		}, "",
	)
	if err != nil {
		t.Error(err)
	}

	if len(agents) != 1 {
		t.Error("Incorrect number of agents found")
	}
}

func TestDisplayAgents(t *testing.T) {
	var capi ClientApi
	err := capi.DisplayAgents([]defs.AgentAdvert{}, false)
	if err != nil {
		t.Error(err)
	}
}
