/*
   ToDD Client API

   Copyright 2015 - Matt Oswalt
*/

package api

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "os"
    "text/template"

    "github.com/mierdin/todd/agent/defs"
    "github.com/mierdin/todd/common"
)

func (capi ClientApi) Agent(conf map[string]string, subargs []string) {

    var templ_final *template.Template
    var url string

    url = fmt.Sprintf("http://%s:%s/v1/agent", conf["host"], conf["port"])

    // If no UUID was provided, get all agents
    if len(subargs) == 0 {

        tmpl, err := template.New("test").Parse(
            `Agent UUID:  {{.Uuid}}
Expires:  {{.Expires}}
Facts Summary: {{.FactSummary}}
Collector Summary: {{.CollectorSummary}}` + "\n")

        if err != nil {
            panic(err)
        } else {
            templ_final = tmpl
        }

    } else {
        url = fmt.Sprintf("%s%s%s", url, "?uuid=", subargs[0])

        // TODO(moswalt): if nothing found, API should return either null or empty slice, and client should handle this
        tmpl, err := template.New("test").Parse(
            `Agent UUID:  {{.Uuid}}
Expires:  {{.Expires}}
Collector Summary: {{.CollectorSummary}}
Facts:
{{.PPFacts}}` + "\n")

        if err != nil {
            panic(err)
        } else {
            templ_final = tmpl
        }
    }

    // Build the request
    req, err := http.NewRequest("GET", url, nil)
    common.FailOnError(err, "Failed to retrieve data from server")

    // Send the request via a client
    client := &http.Client{}
    resp, err := client.Do(req)
    common.FailOnError(err, "Failed to retrieve data from server")

    // Defer the closing of the body
    defer resp.Body.Close()
    // Read the content into a byte array
    body, err := ioutil.ReadAll(resp.Body)
    common.FailOnError(err, "Failed to retrieve data from server")

    // Marshal API data into object
    var records []defs.AgentAdvert
    err = json.Unmarshal(body, &records)
    common.FailOnError(err, "Failed to retrieve data from server")

    // Output retrieved data
    fmt.Println(common.HR)
    for i := range records {
        err = templ_final.Execute(os.Stdout, records[i])
        if err != nil {
            panic(err)
        }

        fmt.Println(common.HR)
    }

}
