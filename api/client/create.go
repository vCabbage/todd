/*
    ToDD Client API Calls for "todd create"

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/Mierdin/todd/blob/master/LICENSE
*/

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"

	"github.com/Mierdin/todd/server/objects"
)

// Create is responsible for pushing a ToDD object to the server for eventual storage in whatever database is being used
// It will send a ToddObject rendered as JSON to the "createobject" method of the ToDD API
func (capi ClientApi) Create(conf map[string]string, yamlFileName string) {

	// See if anything is being passed via stdin
	fi, err := os.Stdin.Stat()
	if err != nil {
		panic(err)
	}

	var yamlDef []byte
	// If stdin is populated, read from that
	if fi.Size() > 0 {
		yamlDef, err = ioutil.ReadAll(os.Stdin)
		if err != nil {
			panic(err)
		}
	} else {

		// Quit if there's nothing on stdin, and there's no arg either
		if yamlFileName == "" {
			fmt.Println("Please provide definition file")
			os.Exit(1)
		}

		// Read YAML file
		filename, _ := filepath.Abs(fmt.Sprintf("./%s", yamlFileName))
		yamlDef, err = ioutil.ReadFile(filename)
		if err != nil {
			fmt.Println("Unable to parse YAML")
			os.Exit(1)
		}
	}

	// Unmarshal YAML file into a BaseObject so we can peek into the metadata
	var baseobj objects.BaseObject
	err = yaml.Unmarshal(yamlDef, &baseobj)
	if err != nil {
		panic(err)
	}

	// finalobj represents the object being created, regardless of type.
	// ToddObject is an interface that satisfies all ToDD objects
	var finalobj objects.ToddObject

	switch baseobj.Type {
	case "group":
		var group_obj objects.GroupObject
		err = yaml.Unmarshal(yamlDef, &group_obj)
		if err != nil {
			panic(err)
		}
		finalobj = group_obj
	case "testrun":
		var testrun_obj objects.TestRunObject
		err = yaml.Unmarshal(yamlDef, &testrun_obj)
		if err != nil {
			panic(err)
		}

		if testrun_obj.Spec.TargetType == "group" {

			// We need to do a quick conversion because JSON does not support non-string
			// keys, and would reject this during Marshal if we don't.
			stringified_map := make(map[string]string)
			for k, v := range testrun_obj.Spec.Target.(map[interface{}]interface{}) {
				stringified_map[k.(string)] = v.(string)
			}
			testrun_obj.Spec.Target = stringified_map

		}

		finalobj = testrun_obj

	default:
		fmt.Println("Invalid object type provided")
		os.Exit(1)
	}

	// Marshal the final object into JSON
	json_str, err := json.Marshal(finalobj)
	if err != nil {

		panic(err)
	}

	// Construct API request, and send POST to server for this object
	var url string
	url = fmt.Sprintf("http://%s:%s/v1/object/create", conf["host"], conf["port"])

	var jsonByte = []byte(json_str)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	if err != nil {
		panic(err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// Print a regular OK message if object was written successfully - else print some debug info
	if resp.Status == "200 OK" {
		fmt.Println("[OK]")
	} else {
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("response Body:", string(body))
		os.Exit(1)
	}

}
