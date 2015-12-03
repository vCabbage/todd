/*
   ToDD API

   Copyright 2015 - Matt Oswalt
*/

package api

import (
    "encoding/json"
    "fmt"
    // "github.com/mierdin/todd/agent/defs"
    "github.com/mierdin/todd/config"
    "github.com/mierdin/todd/db"
    "net/http"
)

func (tapi ToDDApi) ListObjects(w http.ResponseWriter, r *http.Request) {

}

func (tapi ToDDApi) NewObject(w http.ResponseWriter, r *http.Request) {

    type ToDDObj struct {
        Label   string              `json:"Label"`
        Matches []map[string]string `json:"Matches"`
    }

    decoder := json.NewDecoder(r.Body)
    var gf GroupFile
    err := decoder.Decode(&gf)
    if err != nil {
        panic(err)
    }
    fmt.Println(gf.Group)
    //

    // TODO(mierdin): better config
    cfg := config.GetConfig("/etc/server_config.cfg")
    var tdb = db.NewToddDB(cfg)

    tdb.DatabasePackage.SetGroupFile(gf.Group, gf.Matches)

    // response, err := json.MarshalIndent(agent_list, "", "  ")

    // if err != nil {
    //     panic(err)
    // }

    // fmt.Fprint(w, string(response))
}
