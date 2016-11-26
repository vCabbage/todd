/*
    ToDD Client API Calls for "todd agents"

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"text/template"

	"text/tabwriter"

	"github.com/toddproject/todd/agent/defs"
	"github.com/toddproject/todd/hostresources"
)

// Agents will query the ToDD server for a list of currently registered agents, and will display
// a list of them to the user. Optionally, the user can provide a subargument containing the UUID of
// a registered agent, and this function will output more detailed information about that agent.
func (c *ClientAPI) Agents(agentUUID string) ([]defs.AgentAdvert, error) {
	url := c.baseURL + "/agent"
	if agentUUID != "" {
		url = fmt.Sprintf("%s?uuid=%s", url, agentUUID)
	}

	resp, err := c.http.Get(url)
	if err != nil {
		return nil, err
	}
	defer io.Copy(ioutil.Discard, resp.Body) // Ensure fully read so client can be reused
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New(resp.Status)
	}

	// Marshal API data into object
	var agents []defs.AgentAdvert
	err = json.NewDecoder(resp.Body).Decode(&agents)

	return agents, err
}

var detailTemplate = template.Must(template.New("test").Parse(
	`{{range .}}Agent UUID:  {{.UUID}}
Expires:  {{.Expires}}
Collector Summary: {{.CollectorSummary}}
Facts:
{{.PPFacts}}
{{end}}`))

// DisplayAgents is responsible for displaying a set of Agents to the terminal
func (c *ClientAPI) DisplayAgents(agents []defs.AgentAdvert, detail bool) error {
	switch {
	case len(agents) == 0, agents[0].UUID == "":
		fmt.Println("No agents found.")
		return nil
	case detail:
		// TODO(moswalt): if nothing found, API should return either null or empty slice, and client should handle this

		// Output retrieved data
		return detailTemplate.Execute(os.Stdout, agents)
	}

	// Format in tab-separated columns with a tab stop of 8.
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintln(w, "UUID\tEXPIRES\tADDR\tFACT SUMMARY\tCOLLECTOR SUMMARY")

	for i := range agents {
		fmt.Fprintf(
			w,
			"%s\t%s\t%s\t%s\t%s\n",
			hostresources.TruncateID(agents[i].UUID),
			agents[i].Expires,
			agents[i].DefaultAddr,
			agents[i].FactSummary(),
			agents[i].CollectorSummary(),
		)
	}
	fmt.Fprintln(w)
	w.Flush()

	return nil
}
