/*
    ToDD API - manages agents

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Mierdin/todd/agent/defs"
)

func (tapi ToDDApi) Agent(w http.ResponseWriter, r *http.Request) {
	agentList, err := tapi.tdb.GetAgents()
	if err != nil {
		http.Error(w, "Internal Error", 500)
		return
	}

	// Make sure UUID string is provided
	if uuid := r.URL.Query().Get("uuid"); uuid != "" {
		for i := 0; len(agentList); i++ {
			if strings.HasPrefix(agentList[i].Uuid, uuid[0]) {
				agentList = []defs.AgentAdvert{agentList[i]}
				break
			}
		}
	}

	// If there are no agents, return an empty slice, not a nil slice - this
	// prevents this API from returning "null"
	if agentList == nil {
		agentList = []defs.AgentAdvert{}
	}

	response, err := json.MarshalIndent(agentList, "", "  ")
	if err != nil {
		http.Error(w, "Internal Error", 500)
		return
	}

	w.Write(response)
}
