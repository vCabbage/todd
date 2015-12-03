/*
   ToDD comms functions

    This file holds the infrastructure for agent-server communication abstractions in ToDD.

   Copyright 2015 - Matt Oswalt
*/

package comms

import (
    "github.com/mierdin/todd/agent/defs"
    "github.com/mierdin/todd/config"
)

// CommsPackage will ensure that whatever specific comms struct is loaded at compile time will support
// all of the necessary features/functions that we need to make ToDD work. In short, this interface
// represents a list of things that the server and agents do on the message queue.
type CommsPackage interface {

    // TODO(mierdin) best way to document interface or function args?

    // (agent advertisement to advertise)
    AdvertiseAgent(defs.AgentAdvert)

    // (map of collectors:hashes)
    ListenForAgent(map[string]string)

    // (uuid)
    ListenForFactRemediationNotice(string)

    // (uuid, remediations)
    SendFactRemediationNotice(string, []string)
}

// toddComms is a struct to hold anything that satisfies the CommsPackage interface
type toddComms struct {
    CommsPackage
}

// NewToDDComms will create a new instance of toddComms, and load the desired
// CommsPackage-compatible comms package into it.
func NewToDDComms(cfg config.Config) *toddComms {

    // Create comms instance
    var tc toddComms

    // Set CommsPackage to desired package
    // TODO(mierdin): Need to make this dynamic based on config
    var rmq RabbitMQComms
    rmq.config = cfg
    tc.CommsPackage = rmq

    return &tc
}
