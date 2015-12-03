/*
   Configuration services

   Copyright 2015 - Matt Oswalt
*/

package config

import (
    "github.com/mierdin/todd/common"
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
    Facts struct {
        CollectorDir string
        IP           string
        Port         string
    }
}

func GetConfig(cfgpath string) Config {
    var cfg Config

    // TODO: Make the config file a command-line argument

    err := gcfg.ReadFileInto(&cfg, cfgpath)
    common.FailOnError(err, "Error retrieving configuration")

    return cfg
}
