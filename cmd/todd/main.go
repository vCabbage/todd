/*
    ToDD Client - Primary entrypoint

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package main

import (
	"fmt"
	"os"

	cli "github.com/codegangsta/cli"
	"github.com/toddproject/todd"
	capi "github.com/toddproject/todd/api/client"
	"github.com/toddproject/todd/server/objects"
)

func main() {
	app := cli.NewApp()
	app.Name = "todd"
	app.Version = todd.Version
	app.Usage = "A highly extensible framework for distributed testing on demand"

	// global level flags
	var host string
	var port int
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "H, host",
			Usage:       "ToDD server hostname",
			Value:       "localhost",
			Destination: &host,
		},
		cli.IntFlag{
			Name:        "P, port",
			Usage:       "ToDD server API port",
			Value:       8080,
			Destination: &port,
		},
	}

	// Configure clientAPI before other commands are run
	var clientAPI *capi.ClientAPI
	app.Before = func(c *cli.Context) error {
		clientAPI = capi.New(c.String("host"), c.Int("port"))
		return nil
	}

	// ToDD Commands
	app.Commands = []cli.Command{

		// "todd agents ..."
		{
			Name:  "agents",
			Usage: "Show ToDD agent information",
			Action: func(c *cli.Context) {
				agents, err := clientAPI.Agents(c.Args().Get(0))
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
				err = clientAPI.DisplayAgents(agents, !(c.Args().Get(0) == ""))
				if err != nil {
					fmt.Println("Problem displaying agents (client-side)")
					os.Exit(1)
				}
			},
		},

		// "todd create ..."
		{
			Name:  "create",
			Usage: "Create ToDD object (group, testrun, etc.)",
			Action: func(c *cli.Context) {
				err := clientAPI.Create(c.Args().Get(0))
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			},
		},

		// "todd delete ..."
		{
			Name:  "delete",
			Usage: "Delete ToDD object",
			Action: func(c *cli.Context) {
				objType, objLabel := c.Args().Get(0), c.Args().Get(1)
				err := clientAPI.Delete(objType, objLabel)
				if err != nil {
					fmt.Printf("ERROR: %s\n", err)
					fmt.Println("(Are you sure you provided the right object type and/or label?)")
					os.Exit(1)
				}
			},
		},

		// "todd groups ..."
		{
			Name:  "groups",
			Usage: "Show current agent-to-group mappings",
			Action: func(c *cli.Context) {
				err := clientAPI.Groups()
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			},
		},

		// "todd objects ..."
		{
			Name:  "objects",
			Usage: "Show information about installed group objects",
			Action: func(c *cli.Context) {
				err := clientAPI.Objects(c.Args().Get(0))
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			},
		},

		// "todd run ..."
		{
			Name: "run",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name:  "j",
					Usage: "Output test data for this testrun when finished",
				},
				cli.BoolFlag{
					Name:  "y",
					Usage: "Skip confirmation and run referenced testrun immediately",
				},
				cli.StringFlag{
					Name:  "source-group",
					Usage: "The name of the source group",
				},
				cli.StringFlag{
					Name:  "source-app",
					Usage: "The app to run for this test",
				},
				cli.StringFlag{
					Name:  "source-args",
					Usage: "Arguments to pass to the testlet",
				},
			},
			Usage: "Execute an already uploaded testrun object",
			Action: func(c *cli.Context) {
				overrides := objects.SourceOverrides{
					Group: c.String("source-group"),
					App:   c.String("source-app"),
					Args:  c.String("source-args"),
				}
				name := c.Args().Get(0)
				err := clientAPI.Run(overrides, name, c.Bool("j"), c.Bool("y"))
				if err != nil {
					fmt.Println(err)
					os.Exit(1)
				}
			},
		},
	}

	app.Run(os.Args)
}
