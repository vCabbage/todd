/*
    ToDD tsdbPackage implementation for influxdb

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package tsdb

import (
	"fmt"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	influx "github.com/influxdata/influxdb/client/v2"

	"github.com/Mierdin/todd/config"
)

// newInfluxDB is a factory function that produces a new instance of influxDB with the configuration
// loaded and ready to be used.
func newInfluxDB(cfg config.Config) *influxDB {
	var ifdb influxDB
	ifdb.config = cfg
	return &ifdb
}

type influxDB struct {
	config config.Config
}

func (ifdb influxDB) WriteData(testUuid, testRunName, groupName string, testData map[string]map[string]map[string]string) {
	// Make client
	c, _ := influx.NewHTTPClient(influx.HTTPConfig{
		Addr: fmt.Sprintf("http://%s:%s", ifdb.config.TSDB.IP, ifdb.config.TSDB.Port),
	})

	// Create a new point batch
	bp, _ := influx.NewBatchPoints(influx.BatchPointsConfig{
		Database:  "todd_metrics",
		Precision: "s",
	})

	for agentUuid, agentData := range testData {

		for targetAddress, metrics := range agentData {

			// Create a point and add to batch
			tags := map[string]string{
				"agent":       agentUuid,
				"target":      targetAddress,
				"sourceGroup": groupName,
				"testUuid":    testUuid,
			}

			// Convert our metrics to float and insert into influx fields
			fields := make(map[string]interface{})
			for k, v := range metrics {
				float_v, _ := strconv.ParseFloat(v, 64)
				fields[k] = float_v
			}
			pt, _ := influx.NewPoint(fmt.Sprintf("testrun-%s", testRunName), tags, fields, time.Now())
			bp.AddPoint(pt)

		}

	}

	// Write the batch
	c.Write(bp)

	log.Infof("Wrote test data for %s to influxdb", testUuid)
}
