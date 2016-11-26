/*
    ToDD Client API Calls for "todd create"

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/pkg/errors"
	"github.com/toddproject/todd/server/objects"
)

// Create is responsible for pushing a ToDD object to the server for eventual storage in whatever database is being used
// It will send a ToddObject rendered as JSON to the "createobject" method of the ToDD API
func (c *ClientAPI) Create(yamlFileName string) error {
	// Pull YAML from either stdin or from the filename if stdin is empty
	yamlDef, err := getYAMLDef(yamlFileName)
	if err != nil {
		return err
	}

	// Unmarshal YAML file into a BaseObject so we can peek into the metadata
	var baseObj objects.BaseObject
	err = yaml.Unmarshal(yamlDef, &baseObj)
	if err != nil {
		return errors.New("YAML file not in correct format")
	}

	// finalobj represents the object being created, regardless of type.
	// ToddObject is an interface that satisfies all ToDD objects
	var finalObj objects.ToddObject
	switch baseObj.Type {
	case "group":
		finalObj = &objects.GroupObject{}
	case "testrun":
		finalObj = &objects.TestRunObject{}
	default:
		return errors.New("Invalid object type provided")
	}
	err = yaml.Unmarshal(yamlDef, finalObj)
	if err != nil {
		return errors.New("Testrun YAML object not in correct format")
	}

	// Marshal the final object into JSON
	var buf bytes.Buffer
	err = json.NewEncoder(&buf).Encode(finalObj)
	if err != nil {
		return errors.New("Problem marshalling the final object into JSON")
	}

	// Construct API request, and send POST to server for this object
	url := c.baseURL + "/object/create"

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return errors.Wrap(err, "creating request")
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return errors.Wrap(err, "sending request")
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	// Print a regular OK message if object was created successfully - else print the HTTP status code
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	fmt.Println("[OK]")

	return nil
}

// getYAMLDef reads YAML from either stdin or from the filename if stdin is empty
func getYAMLDef(yamlFileName string) ([]byte, error) {
	// If stdin is populated, read from that
	if stat, err := os.Stdin.Stat(); err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		return ioutil.ReadAll(os.Stdin)
	}

	// Quit if there's nothing on stdin, and there's no arg either
	if yamlFileName == "" {
		return nil, errors.New("Object definition file not provided - please provide via filename or stdin")
	}

	// Read YAML file
	return ioutil.ReadFile(yamlFileName)
}
