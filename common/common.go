/*
   Common ToDD functions

   Copyright 2015 - Matt Oswalt

   Common functions to be shared among various packages
*/

package common

import (
	"fmt"

	log "github.com/mierdin/todd/Godeps/_workspace/src/github.com/Sirupsen/logrus"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
		panic(fmt.Sprintf("%s: %s", msg, err))
	}
}

func WarnOnError(err error, msg string) {
	if err != nil {
		log.Errorf("%s: %s", msg, err)
	}
}
