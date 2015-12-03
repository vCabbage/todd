/*
   grouping

   Author: Matt Oswalt
*/

package grouping

import (
	"time"

	log "github.com/mierdin/todd/Godeps/_workspace/src/github.com/Sirupsen/logrus"
	toddapi "github.com/mierdin/todd/api/server"
	"github.com/mierdin/todd/common"
	"github.com/mierdin/todd/comms"
	"github.com/mierdin/todd/config"
	"github.com/mierdin/todd/db"
	"github.com/mierdin/todd/facts"
)

func init() {
	// TODO(moswalt): Implement configurable loglevel in server and agent
	log.SetLevel(log.DebugLevel)
}

func main() {

	//TODO (moswalt): Need to make this configurable
	cfg := config.GetConfig("/etc/server_config.cfg")

	// Start serving collectors, and retrieve map of names and hashes
	fact_collectors := facts.ServeFactCollectors(cfg)

	// Perform database initialization tasks
	var tdb = db.NewToddDB(cfg)
	tdb.DatabasePackage.Init()

	// Initialize API
	var tapi toddapi.ToDDApi
	go tapi.Start(cfg)

	// Start listening for
	var tc = comms.NewToDDComms(cfg)
	go tc.CommsPackage.ListenForAgent(fact_collectors)

	log.Infof("ToDD server %s. Press any key to exit...\n", common.VERSION)

	// Sssh, sssh, only dreams now....
	for {
		time.Sleep(time.Second * 10)
	}
}
