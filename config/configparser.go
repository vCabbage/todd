/*
    ToDD Configuration

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package config

import (
	log "github.com/Sirupsen/logrus"
	"gopkg.in/gcfg.v1"
)

type Config struct {
	API struct {
		Host string
		Port string
	}
	AMQP struct {
		User     string
		Password string
		Host     string
		Port     string
	}
	Comms struct {
		Plugin string
	}
	Assets struct {
		IP   string
		Port string
	}
	DB struct {
		IP     string
		Port   string
		Plugin string
	}
	TSDB struct {
		IP     string
		Port   string
		Plugin string
	}
	Testing struct {
		Timeout int
	}
	Grouping struct {
		Interval int // seconds
	}
	LocalResources struct {
		DefaultInterface string
		OptDir           string
		IPAddrOverride   string
	}
}

func GetConfig(cfgpath string) Config {
	var cfg Config

	err := gcfg.ReadFileInto(&cfg, cfgpath)
	if err != nil {
		log.Error("Error retrieving configuration")
		panic("Error retrieving configuration")
	}

	return cfg
}
