/*
	ToDD tasks (agent communication)

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	"github.com/toddproject/todd/agent/cache"
	"github.com/toddproject/todd/config"
)

// Task is an interface to define task behavior This is used for functions like those in comms
// that need to work with all specific task types, not one specific task.
type Task interface {

	// TODO(mierdin): This works but is a little "meh". Basically, each task has to have a "Run()" function, as enforced
	// by this interface. If the task needs some additional data, it gets these through struct properties. This works but
	// doesn't quite feel right. Come back to this and see if there's a better way.
	Run(*config.Config, *cache.AgentCache, Responder) error
}

// Responder is a function which a task can use to send a response.
type Responder func(interface{}) error

// BaseTask is a struct that is intended to be embedded by specific task structs. Both of these in conjunction
// are used primarily to house the JSON message for passing tasks over the comms package (i.e. message queue), but may also contain important
// dependencies of the task, such as an HTTP handler.
type BaseTask struct {
	Type string `json:"type"`
}
