/*
   ToDD API

   Copyright 2015 - Matt Oswalt
*/

package api

import (
    "fmt"
    "github.com/mierdin/todd/common"
    "github.com/mierdin/todd/config"
    "net/http"

    log "github.com/mierdin/todd/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

type ToDDApi struct{}

func (tapi ToDDApi) Start(cfg config.Config) {

    // TODO(mierdin): Would like to auto-discover API methods present in this package instead
    // of statically listing them here
    // TODO(mierdin): make version dynamic
    http.HandleFunc("/v1/agent", tapi.Agent)
    http.HandleFunc("/v1/listobjects", tapi.GroupFiles)
    http.HandleFunc("/v1/createobject", tapi.NewObject)

    serve_url := fmt.Sprintf("%s:%s", cfg.API.Host, cfg.API.Port)

    log.Infof("Serving ToDD Server API at: %s\n", serve_url)
    err := http.ListenAndServe(serve_url, nil)
    common.FailOnError(err, "Error starting API")
}
