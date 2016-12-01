/*
    ToDD Object API

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package objects

import (
	"encoding/json"

	"github.com/pkg/errors"
)

// BaseObject is the "base" struct for all objects in ToDD. All metadata shared by all todd objects should be stored here.
// Any structs representing todd objects should embed this struct.
//
// Generally, since the "Type" field is present in the "parent", or embedded struct, it is appropriate to first unmarshal YAML
// or JSON into the base "ToddObject" struct first, to determine type. Then, this value is used to determine which "child" struct
// should be used to parse the entire structure, using a "switch" statement, we'll call a "type switch".
//
// The process for adding a new object type in ToDD should be done in 3 places:
//
// - Here, a new struct that embeds ToddObject should be added.
// - The "switch" statement in the ParseToddObject and ParseToddObjects functions below should be augmented to look for the new object type for an incoming YAML file
// - The "Create" function in the client api contains a "type switch" - update this.
//
// It should be mentioned that the steps above only add a new object type to the various infrastructure
// elements of ToDD - it does not automatically give those objects meaning.
type BaseObject struct {
	Label string `json:"label" yaml:"label"`
	Type  string `json:"type" yaml:"type"`
}

// ToddObject is an interface that describes all ToDD Objects. You will see this interface referenced in all
// functions that get, set, or delete ToDD objects.
type ToddObject interface {
	GetType() string
	GetLabel() string
	GetSpec() string
}

// GetType is a simple function to return the "Type" attribute of a BaseObject struct (or any struct that embeds it).
// It is necessary for portions of the code that only have a handle on the ToddObject interface, and need to access these properties
func (b BaseObject) GetType() string {
	return b.Type
}

// GetLabel is a simple function to return the "Label" attribute of a BaseObject struct (or any struct that embeds it).
// It is necessary for portions of the code that only have a handle on the ToddObject interface, and need to access these properties
func (b BaseObject) GetLabel() string {
	return b.Label
}

// ParseToddObject centralizes the logic of generating a specific ToDD object off of an existing base struct. Whenever you need to parse
// a specific ToDD object in the codebase, you should first Unmarshal into the BaseObject struct. Once you have a handle on that, run this struct method, and
// pass in the original JSON (the "obj_json" param). This will identify the specific type and return that struct to you.
// TODO(mierdin): Need a better name for this. Change it, then make sure you're replacing all instances of the old name in this file (in several comments)
func (b *BaseObject) ParseToddObject(objJSON []byte) (ToddObject, error) {
	// The following chunk of code is designed to re-parse the JSON structure of a ToDD object, now that we know what type it is
	// (this "Type" field is present in the base BaseObject struct).
	var tObj ToddObject
	switch b.Type {
	case "group":
		var obj GroupObject
		err := json.Unmarshal(objJSON, &obj)
		if err != nil {
			return nil, err
		}
		tObj = obj
	case "testrun":
		var obj TestRunObject
		err := json.Unmarshal(objJSON, &obj)
		if err != nil {
			return nil, err
		}
		tObj = obj
	default:
		return nil, errors.Errorf("unknown object type: %s", b.Type)
	}

	return tObj, nil
}

// ParseToddObjects is similar to the struct ParseToddObject method but is intended to be used when you know the JSON houses a slice/list of Todd Objects.
// As a result, it will also return a slice of ToddObjects.
// TODO(mierdin): This function is still fairly inflexible, as it assumes that all objects in the JSON are of the same type. Currently, this is a safe bet because
// the objects API doesn't currently support returning all objects, so they will all be the same type. However, if that changes in the future, this function
// will have to be refactored.
func ParseToddObjects(objJSON []byte) ([]ToddObject, error) {
	// Marshal API data into object
	var records []BaseObject
	err := json.Unmarshal(objJSON, &records)
	if err != nil {
		return nil, err
	}

	retSlice := make([]ToddObject, len(records))

	// Return an empty slice since there were no objects in the original JSON
	if len(records) < 1 {
		return retSlice, nil
	}

	switch objType := records[0].GetType(); objType {
	case "group":
		var objs []GroupObject
		err := json.Unmarshal(objJSON, &objs)
		if err != nil {
			return nil, err
		}

		for i, obj := range objs {
			retSlice[i] = obj
		}
	case "testrun":
		var objs []TestRunObject
		err := json.Unmarshal(objJSON, &objs)
		if err != nil {
			return nil, err
		}

		for i, obj := range objs {
			retSlice[i] = obj
		}
	default:
		return nil, errors.Errorf("unknown object type: %s", objType)
	}

	return retSlice, nil
}
