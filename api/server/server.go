/*
    ToDD API

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/toddproject/todd/db"
	"github.com/toddproject/todd/server"

	"github.com/toddproject/todd/config"
)

// ServerAPI contains necessary components for server handlers.
type ServerAPI struct {
	cfg    config.Config
	tdb    db.DatabasePackage
	Server *server.Server
}

// Start configured ServerAPI, registers handlers, and starts listening only
// the configured host and port.
//
// Start blocks until http.ListenAndServe returns.
func (s *ServerAPI) Start(cfg config.Config) error {
	s.cfg = cfg

	tdb, err := db.NewToddDB(s.cfg)
	if err != nil {
		return err
	}

	s.tdb = tdb

	// TODO(mierdin): This needs a lot of work. Not only is the version very static
	// (which is okay for now until we hit a new version)
	// but the static routing is gross. Make this not suck
	http.HandleFunc("/v1/agent", s.Agent)
	http.HandleFunc("/v1/groups", s.Groups)
	http.HandleFunc("/v1/object/list", s.ListObjects)
	http.HandleFunc("/v1/object/group", s.ListObjects)
	http.HandleFunc("/v1/object/testrun", s.ListObjects)
	http.HandleFunc("/v1/object/create", s.CreateObject)
	http.HandleFunc("/v1/object/delete", s.DeleteObject)
	http.HandleFunc("/v1/testrun/run", s.Run)
	http.HandleFunc("/v1/testdata", s.TestData)

	serveURL := fmt.Sprintf("%s:%s", s.cfg.API.Host, s.cfg.API.Port)

	log.Infof("Serving ToDD Server API at: %s\n", serveURL)
	return http.ListenAndServe(serveURL, nil)
}

// writeJSON writes obj to w as JSON.
//
// An error will be written to w if JSON encoding fails.
func writeJSON(w http.ResponseWriter, obj interface{}) {
	err := json.NewEncoder(w).Encode(obj)
	if err != nil {
		writeError(w, err)
	}
}

// writeError logs the error and sends a 500 to the client.
func writeError(w http.ResponseWriter, err error) {
	log.Errorln(err)
	http.Error(w, "Internal Error", 500)
}
