/*
    ToDD API

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"fmt"
	"net/http"
	"os"

	log "github.com/Sirupsen/logrus"

	"github.com/Mierdin/todd/config"
)

type ToDDApi struct {
	cfg config.Config
}

func (tapi ToDDApi) Start(cfg config.Config) {

	tapi.cfg = cfg

	// TODO(mierdin): This needs a lot of work. Not only is the version very static
	// (which is okay for now until we hit a new version)
	// but the static routing is gross. Make this not suck
	http.HandleFunc("/v1/agent", tapi.Agent)
	http.HandleFunc("/v1/groups", tapi.Groups)
	http.HandleFunc("/v1/object/list", tapi.ListObjects)
	http.HandleFunc("/v1/object/group", tapi.ListObjects)
	http.HandleFunc("/v1/object/testrun", tapi.ListObjects)
	http.HandleFunc("/v1/object/create", tapi.CreateObject)
	http.HandleFunc("/v1/object/delete", tapi.DeleteObject)
	http.HandleFunc("/v1/testrun/run", tapi.Run)
	http.HandleFunc("/v1/testdata", tapi.TestData)

	serve_url := fmt.Sprintf("%s:%s", tapi.cfg.API.Host, tapi.cfg.API.Port)

	log.Infof("Serving ToDD Server API at: %s\n", serve_url)
	err := http.ListenAndServe(serve_url, nil)
	if err != nil {
		log.Error("Error starting API")
		os.Exit(1)
	}
}
