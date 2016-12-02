/*
    ToDD tsdbPackage implementation for influxdb

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tsdb

import (
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	influx "github.com/influxdata/influxdb/client/v2"
	"github.com/pkg/errors"

	"github.com/toddproject/todd/config"
)

func init() {
	register("influxdb", newInfluxDB)
}

// newInfluxDB is a factory function that produces a new instance of influxDB with the configuration
// loaded and ready to be used.
func newInfluxDB(cfg *config.Config) (TSDB, error) {
	db := &influxDB{config: cfg}

	if err := db.init(); err != nil {
		return nil, err
	}

	return db, nil
}

type influxDB struct {
	config *config.Config
	client influx.Client
}

func (db *influxDB) init() error {
	client, err := influx.NewHTTPClient(influx.HTTPConfig{
		Addr: fmt.Sprintf("http://%s:%s", db.config.TSDB.Host, db.config.TSDB.Port),
	})
	if err != nil {
		return errors.Wrap(err, "creating client")
	}

	db.client = client

	_, err = client.Query(influx.Query{
		Command:  "CREATE DATABASE " + db.config.TSDB.DatabaseName,
		Database: db.config.TSDB.DatabaseName,
	})
	if err != nil {
		client.Close()
		return errors.Wrap(err, "creating database")
	}

	return nil
}

func (db *influxDB) Close() error {
	return db.client.Close()
}

// WriteData will write the resulting testrun data to influxdb as a batch of points - containing
// important information like metrics and which agent reported them.
func (db *influxDB) WriteData(testUUID, testRunName, groupName string, testData map[string]map[string]map[string]interface{}) error {
	// Create a new point batch
	bp, err := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  db.config.TSDB.DatabaseName,
		Precision: "s",
	})
	if err != nil {
		return errors.Wrap(err, "creating batch points")
	}

	// Need to publish data from all of the agents that took part in this test
	for agentUUID, agentData := range testData {
		// Also need to differentiate between the various target that these agents tested against
		for targetAddress, metrics := range agentData {
			// Create a point and add to batch
			tags := map[string]string{
				"agent":       agentUUID,
				"target":      targetAddress,
				"sourceGroup": groupName,
				"testUuid":    testUUID,
			}

			// Insert into influx fields
			fields := make(map[string]interface{})
			for k, v := range metrics {
				fields[k] = v
			}
			pt, err := influx.NewPoint("testrun-"+testRunName, tags, fields, time.Now())
			if err != nil {
				return errors.Wrap(err, "creating point")
			}

			bp.AddPoint(pt)
		}
	}

	// Write the batch
	err = db.client.Write(bp)
	if err != nil {
		return errors.Wrap(err, "writing to DB")
	}

	log.Infof("Wrote test data for %s to influxdb", testUUID)

	return nil
}
