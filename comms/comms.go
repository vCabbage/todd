/*
    ToDD comms functions

    This file holds the infrastructure for agent-server communication abstractions in ToDD.

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package comms

import (
	"context"
	"fmt"
	"sync"

	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/agent/tasks"
	"github.com/toddproject/todd/config"
)

// Comms will ensure that whatever specific comms struct is loaded at compile time will support
// all of the necessary features/functions that we need to make ToDD work. In short, this interface
// represents a list of things that the server and agents do on the message queue.
type Comms interface {
	// Used by server
	ListenForAgent(context.Context) (chan []byte, error)
	ListenForResponses(context.Context) (chan []byte, error)
	SendTask(to string, task tasks.Task) error

	// Used by agent
	AdvertiseAgent(defs.AgentAdvert) error
	ListenForTasks(from string, ctx context.Context) (chan []byte, error)
	SendResponse(msg interface{}) error
}

// New will create a new instance of toddComms, and load the desired
// CommsPackage-compatible comms package into it.
func New(cfg *config.Config) (Comms, error) {
	construct, ok := packages[cfg.Comms.Plugin]
	if !ok {
		return nil, fmt.Errorf("Invalid comms plugin %q in config file", cfg.Comms.Plugin)
	}

	return construct(cfg)
}

type constructor func(*config.Config) (Comms, error)

var (
	packagesMu sync.Mutex
	packages   = make(map[string]constructor)
)

func register(name string, c constructor) {
	packagesMu.Lock()
	packages[name] = c
	packagesMu.Unlock()
}
