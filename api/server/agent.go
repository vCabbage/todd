/*
    ToDD API - manages agents

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"net/http"
	"strings"

	"github.com/toddproject/todd/agent/defs"
)

// Agent returns a list of agents.
//
// Agents can be filtered by providing the uuid query param ("?uuid=1234")
func (s *ServerAPI) Agent(w http.ResponseWriter, r *http.Request) {
	agentList, err := s.tdb.GetAgents()
	if err != nil {
		writeError(w, err)
		return
	}

	// Make sure UUID string is provided
	if uuid := r.URL.Query().Get("uuid"); uuid != "" {
		// Let's use the full list so we can identify the right agent if the user specified a short
		for i := range agentList {
			if strings.HasPrefix(agentList[i].UUID, uuid) {
				// Replace agentList with first match and break
				agentList = []defs.AgentAdvert{agentList[i]}
				break
			}
		}
	}

	writeJSON(w, agentList)
}
