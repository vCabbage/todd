/*
   ToDD Client API

   Copyright 2015 - Matt Oswalt
*/

package api

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "path/filepath"

    "gopkg.in/yaml.v2"
)

type GroupFile struct {
    Group   string              `yaml:"Group"`
    Matches []map[string]string `yaml:"Matches"`
}

func (capi ClientApi) GroupFile(conf map[string]string, subargs []string) {

    // If no subarg was provided, just get current groups and display
    if len(subargs) == 0 {

        //TODO(mierdin): Implement this

    } else {

        // need to see if subargs[0] is a valid yml file, and import it to JSON here

        filename, _ := filepath.Abs(fmt.Sprintf("./%s", subargs[0]))
        //fmt.Println(filename)
        yamlFile, err := ioutil.ReadFile(filename)

        if err != nil {
            panic(err)
        }

        var gf GroupFile

        err = yaml.Unmarshal(yamlFile, &gf)
        if err != nil {
            panic(err)
        }

        var url string

        url = fmt.Sprintf("http://%s:%s/v1/newgroupfile", conf["host"], conf["port"])

        json_str, err := json.Marshal(gf)

        var jsonByte = []byte(json_str)
        req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
        req.Header.Set("X-Custom-Header", "myvalue")
        req.Header.Set("Content-Type", "application/json")

        client := &http.Client{}
        resp, err := client.Do(req)
        if err != nil {
            panic(err)
        }
        defer resp.Body.Close()

        fmt.Println("response Status:", resp.Status)
        fmt.Println("response Headers:", resp.Header)
        body, _ := ioutil.ReadAll(resp.Body)
        fmt.Println("response Body:", string(body))

    }
}
