/*
   ToDD Client

   Copyright 2015 - Matt Oswalt
*/

package main

import (
    "flag"
    "fmt"
    "os"

    capi "github.com/mierdin/todd/api/client"
    "github.com/mierdin/todd/common"
)

func printVersion() {
    fmt.Printf("ToDD %s\n", common.VERSION)
    os.Exit(0)
}

// Command-line Arguments
var arg_host string
var arg_port string
var arg_version bool
var arg_loglevel string
var arg_help bool

func init() {

    // TODO(moswalt): can you also support shorter flags simultaneously?

    flag.Usage = func() {
        fmt.Print(`Usage: todd-client [OPTIONS] COMMAND [arg...]

    An extensible framework for providing natively distributed testing on demand

    Options:
      --host="localhost"          ToDD server hostname
      --port="8080"               ToDD server port
      --log-level=info            Set the logging level
      --help                	  Print usage
      --version                   Print version information and quit

    Commands:
        agent            Show agent information

    Run 'todd-client COMMAND --help' for more information on a command.`, "\n\n")

        os.Exit(0)
        //TODO (moswalt): implement keyword-specific arguments - may want to move these to another file
    }

    flag.StringVar(&arg_host, "host", "localhost", "ToDD server hostname")
    flag.StringVar(&arg_port, "port", "8080", "ToDD server port")
    flag.BoolVar(&arg_version, "version", false, "Print version information and quit")
    flag.StringVar(&arg_loglevel, "log-level", "info", "Set logging level")
    flag.BoolVar(&arg_help, "h", false, "Print usage")
    flag.Parse()
}

func main() {
    if arg_help {
        flag.Usage()
        os.Exit(0)
    }

    if arg_version {
        printVersion()
        os.Exit(0)
    }

    // Create conf map
    // TODO(moswalt): This may need to be a struct
    conf := map[string]string{}
    conf["host"] = arg_host
    conf["port"] = arg_port

    // Ensure that positional arguments are provided
    if len(flag.Args()) == 0 {
        flag.Usage()
        os.Exit(0)
    } else {

        var clientapi capi.ClientApi

        // Call appropriate function based on positional arguent
        switch flag.Args()[0] {
        case "agent":
            clientapi.Agent(conf, flag.Args()[1:])
        case "groupfile":
            clientapi.GroupFile(conf, flag.Args()[1:])
        }
    }
}

// TODO(moswalt): This is something docker is doing. See if this is useful
// type command struct {
//     name        string
//     description string
// }

// var (
//     toddCommands = []command{
//         {"agent", "Show detailed information about an agent"},
//         {"agentlist", "Build an image from a Dockerfile"},
//     }
// )
