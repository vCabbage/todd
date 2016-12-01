/*
   ToDD API - manages todd objects

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/toddproject/todd/api"
	"github.com/toddproject/todd/server/objects"
)

// ListObjects will query the database layer for a slice of objects based on the parameters provided in the URL.
// The client may ask for a specific type of object by using the "/all" notation after the object keyword in the URL,
// or this may be a specific type, such as "/group".
func (s *ServerAPI) ListObjects(w http.ResponseWriter, r *http.Request) {
	// Derive specific type from URL
	objType := strings.Split(r.URL.Path, "/")[3]

	objectList, err := s.tdb.GetObjects(objType)
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, objectList)
}

// CreateObject will decode a JSON object into a proper ToddObject instance, and send that to the
// database layer to be written persistently.
func (s *ServerAPI) CreateObject(w http.ResponseWriter, r *http.Request) {
	// Read the content into a byte array
	// (we're doing this so we can access the JSON contents more than once)
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		writeError(w, err)
		return
	}

	// Marshal API data into BaseObject
	var baseObj objects.BaseObject
	err = json.Unmarshal(body, &baseObj)
	if err != nil {
		writeError(w, err)
		return
	}

	// Generate a more specific Todd Object based on the JSON data
	finalObj, err := baseObj.ParseToddObject(body)
	if err != nil {
		writeError(w, err)
		return
	}

	err = s.tdb.SetObject(finalObj)
	if err != nil {
		writeError(w, err)
		return
	}
}

// DeleteObject will decode a JSON object from a client request to determine the type and label of the
// object that needs to be deleted. Then, it will send this information to the database layer to delete this object.
func (s *ServerAPI) DeleteObject(w http.ResponseWriter, r *http.Request) {
	var deleteInfo api.DeleteInfo
	err := json.NewDecoder(r.Body).Decode(&deleteInfo)
	if err != nil {
		writeError(w, err)
		return
	}

	err = s.tdb.DeleteObject(deleteInfo.Label, deleteInfo.Type)
	if err != nil {
		writeError(w, err)
		return
	}
}
