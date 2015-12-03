/*
   ToDD API

   Copyright 2015 - Matt Oswalt
*/

package api

import (
    "encoding/json"
    "fmt"
    "github.com/mierdin/todd/agent/defs"
    "github.com/mierdin/todd/config"
    "github.com/mierdin/todd/db"
    "net/http"
)

func (tapi ToDDApi) Agent(w http.ResponseWriter, r *http.Request) {

    // Retrieve query values
    values := r.URL.Query()

    var agent_list []defs.AgentAdvert

    // TODO(mierdin): better config
    cfg := config.GetConfig("/etc/server_config.cfg")
    var tdb = db.NewToddDB(cfg)

    // Make sure UUID string is provided
    if uuid, ok := values["uuid"]; ok {

        // Make sure UUID string actually contains something
        if len(uuid[0]) > 0 {
            agent_list = tdb.DatabasePackage.GetAgents(uuid[0])
        }

    } else { // UUID not provided; get all agents
        agent_list = tdb.DatabasePackage.GetAgents("")
    }

    response, err := json.MarshalIndent(agent_list, "", "  ")

    if err != nil {
        panic(err)
    }

    fmt.Fprint(w, string(response))
}
