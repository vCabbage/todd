/*
    Agent advertisement definitions

	This package houses a few miscellaneous struct and function
	definitions. It's not very Go-idiomatic, so these definitions
	will likely be moved elsewhere in the near future

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package defs

import (
	"encoding/json"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
)

// AgentAdvert is a struct for an Agent advertisement
type AgentAdvert struct {
	UUID           string              `json:"Uuid"`
	DefaultAddr    string              `json:"DefaultAddr"`
	Expires        time.Duration       `json:"Expires"`
	LocalTime      time.Time           `json:"LocalTime"`
	Facts          map[string][]string `json:"Facts"`
	FactCollectors map[string]string   `json:"FactCollectors"`
	Testlets       map[string]string   `json:"Testlets"`
}

// FactSummary produces a string containing a list of facts present in this agent advertisement.
func (a AgentAdvert) FactSummary() string {
	var keys []string

	for k := range a.Facts {
		keys = append(keys, k)
	}

	return strings.Join(keys, ", ")
}

// CollectorSummary produces a string containing a list of available collectors
// indicated by this agent advertisement.
func (a AgentAdvert) CollectorSummary() string {
	var keys []string

	for k := range a.FactCollectors {
		keys = append(keys, k)
	}

	return strings.Join(keys, ", ")
}

// TestletSummary produces a string containing a list of available collectors
// indicated by this agent advertisement.
func (a AgentAdvert) TestletSummary() string {
	var keys []string

	for k := range a.Testlets {
		keys = append(keys, k)
	}

	return strings.Join(keys, ", ")
}

// PPFacts pretty-prints the facts for an agent
func (a AgentAdvert) PPFacts() string {
	retjson, err := json.MarshalIndent(a.Facts, "", "    ")
	if err != nil {
		log.Warn("Error Pretty-Printing Facts JSON")
	}

	return string(retjson)
}
