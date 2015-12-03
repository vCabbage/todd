/*
   Agent advertisement definitions

   Copyright 2015 - Matt Oswalt
*/

package defs

import (
    "bytes"
    "encoding/json"
    "sync"
    "time"

    "github.com/mierdin/todd/common"
)

type AgentRegistry struct {
    Agents map[string]*AgentAdvert
    Mu     sync.Mutex
}

// AgentAdvert is a struct for an Agent advertisement
type AgentAdvert struct {
    Uuid           string            `json:"Uuid"`
    Expires        time.Duration     `json:"Expires"`
    Facts          interface{}       `json:"Facts"`
    FactCollectors map[string]string `json:"FactCollectors"`
}

// FactSummary produces a string containing a list of facts present in this agent advertisement.
func (a AgentAdvert) FactSummary() string {

    var keys []string

    for k, _ := range a.Facts.(map[string]interface{}) {
        keys = append(keys, k)
    }

    var buffer bytes.Buffer

    for i := range keys {
        buffer.WriteString(keys[i])

        // Append a comma if there are more to write
        if i != len(keys)-1 {
            buffer.WriteString(", ")
        }
    }

    return buffer.String()
}

// CollectorSummary produces a string containing a list of available collectors
// indicated by this agent advertisement.
func (a AgentAdvert) CollectorSummary() string {

    var keys []string

    for k, _ := range a.FactCollectors {
        keys = append(keys, k)
    }

    var buffer bytes.Buffer

    for i := range keys {
        buffer.WriteString(keys[i])

        // Append a comma if there are more to write
        if i != len(keys)-1 {
            buffer.WriteString(", ")
        }
    }

    return buffer.String()
}

// JsonPP pretty-prints the facts for an agent
func (a AgentAdvert) PPFacts() string {
    retjson, err := json.MarshalIndent(a, "", "    ")
    common.FailOnError(err, "Error Pretty-Printing Facts JSON")
    return string(retjson)
}

// Struct for a network interface on an agent
// type AgentNic struct {
//     Name      string   `json:"Name"`
//     HwAddr    string   `json:"HwAddr"`
//     IPv4Addrs []string `json:"IPv4Addrs"`
//     IPv6Addrs []string `json:"IPv6Addrs"`

//     //TODO(moswalt): Change to better datatypes
//     //TODO(moswalt): Isolate mask to addtl property? Maybe filtering after the fact is okay
// }
