/*
    ToDD API - manages agent groups

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
)

// Groups returns a list of all groups.
func (s *ServerAPI) Groups(w http.ResponseWriter, r *http.Request) {
	log.Info("Received request for group map")

	groupMap, err := s.tdb.GetGroupMap()
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, groupMap)
}
