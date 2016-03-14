/*
    ToDD API - manages agent groups

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"

	"github.com/Mierdin/todd/db"
)

func (tapi ToDDApi) Groups(w http.ResponseWriter, r *http.Request) {

	log.Info("Received request for group map")

	var tdb = db.NewToddDB(tapi.cfg)

	groupmap := tdb.DatabasePackage.GetGroupMap()

	// If there are no objects, return an empty slice, not a nil slice - this
	// prevents this API from returning "null"
	if groupmap == nil {
		groupmap = map[string]string{}
	}

	response, err := json.MarshalIndent(groupmap, "", "  ")
	if err != nil {
		panic(err)
	}

	fmt.Fprint(w, string(response))
}
