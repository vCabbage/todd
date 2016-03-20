/*
    ToDD API - manages agents

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/Mierdin/todd/agent/defs"
	log "github.com/Sirupsen/logrus"
)

func (tapi ToDDApi) Agent(w http.ResponseWriter, r *http.Request) {

	// Retrieve query values
	values := r.URL.Query()

	agent_list, err := tapi.tdb.GetAgents()
	if err != nil {
		log.Errorln(err)
		http.Error(w, "Internal Error", 500)
		return
	}

	// Make sure UUID string is provided
	if uuid, ok := values["uuid"]; ok {

		// Make sure UUID string actually contains something
		if len(uuid[0]) > 0 {
			// Let's get the full list so we can identify the right agent if the user specified a short
			full_agent_list, _ := tapi.tdb.GetAgents()

			for i := range full_agent_list {
				if strings.HasPrefix(full_agent_list[i].Uuid, uuid[0]) {
					agent, _ := tapi.tdb.GetAgent(full_agent_list[i].Uuid)
					agent_list = append(agent_list, *agent)
					break
				}
			}

		} else { // UUID not provided; get all agents
			agent_list, _ = tapi.tdb.GetAgents()
		}

	} else { // UUID not provided; get all agents
		agent_list, _ = tapi.tdb.GetAgents()
	}

	// If there are no agents, return an empty slice, not a nil slice - this
	// prevents this API from returning "null"
	if agent_list == nil {
		agent_list = []defs.AgentAdvert{}
	}

	response, err := json.MarshalIndent(agent_list, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Fprint(w, string(response))
}
