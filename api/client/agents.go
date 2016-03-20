/*
    ToDD Client API Calls for "todd agents"

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"text/tabwriter"
	"text/template"

	"github.com/Mierdin/todd/agent/defs"
	"github.com/Mierdin/todd/hostresources"
)

// Agents will query the ToDD server for a list of currently registered agents, and will display
// a list of them to the user. Optionally, the user can provide a subargument containing the UUID of
// a registered agent, and this function will output more detailed information about that agent.
func (capi ClientApi) Agents(conf map[string]string, agentUUID string) {

	var url string

	if agentUUID != "" {
		url = fmt.Sprintf("http://%s:%s/v1/agent?uuid=%s", conf["host"], conf["port"], agentUUID)
	} else {
		url = fmt.Sprintf("http://%s:%s/v1/agent", conf["host"], conf["port"])
	}

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}

	// Defer the closing of the body
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		fmt.Printf("Error from API: %d - ", resp.StatusCode)
		io.Copy(os.Stdout, resp.Body)
		return
	}

	// Marshal API data into object
	var records []defs.AgentAdvert
	err = json.NewDecoder(resp.Body).Decode(&records)
	if err != nil {
		panic(err)
	}

	if len(records) == 0 {
		fmt.Println("No agents found.")
		os.Exit(0)
	} else {
		if records[0].Uuid == "" {
			fmt.Println("No agents found.")
			os.Exit(0)
		}
	}

	// If no UUID was provided, get all agents
	if agentUUID == "" {

		w := new(tabwriter.Writer)

		// Format in tab-separated columns with a tab stop of 8.
		w.Init(os.Stdout, 0, 8, 0, '\t', 0)
		fmt.Fprintln(w, "UUID\tEXPIRES\tADDR\tFACT SUMMARY\tCOLLECTOR SUMMARY")

		for i := range records {
			fmt.Fprintf(
				w,
				"%s\t%s\t%s\t%s\t%s\n",
				hostresources.TruncateID(records[i].Uuid),
				records[i].Expires,
				records[i].DefaultAddr,
				records[i].FactSummary(),
				records[i].CollectorSummary(),
			)
		}
		fmt.Fprintln(w)
		w.Flush()

	} else {

		// TODO(moswalt): if nothing found, API should return either null or empty slice, and client should handle this
		tmpl, err := template.New("test").Parse(
			`Agent UUID:  {{.Uuid}}
Expires:  {{.Expires}}
Collector Summary: {{.CollectorSummary}}
Facts:
{{.PPFacts}}` + "\n")

		if err != nil {
			panic(err)
		}

		// Output retrieved data
		for i := range records {
			err = tmpl.Execute(os.Stdout, records[i])
			if err != nil {
				panic(err)
			}
		}
	}

}
