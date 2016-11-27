/*
    ToDD Client API Calls for "todd delete"

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/toddproject/todd/api"
)

// Delete will send a request to remove an existing ToDD Object
func (c *ClientAPI) Delete(objType, objLabel string) error {
	// If insufficient subargs were provided, error out
	if objType == "" || objLabel == "" {
		return errors.New("Error, need to provide type and label (Ex. 'todd delete group datacenter')")
	}

	// anonymous struct to hold our delete info
	deleteInfo := api.DeleteInfo{
		Label: objLabel,
		Type:  objType,
	}

	// Marshal deleteinfo into JSON
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(deleteInfo)
	if err != nil {
		return err
	}

	// Construct API request, and send POST to server for this object
	url := c.baseURL + "/object/delete"

	req, err := http.NewRequest("POST", url, &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	io.Copy(ioutil.Discard, resp.Body)
	resp.Body.Close()

	// Print a regular OK message if object was written successfully - else print the HTTP status code
	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}
	fmt.Println("[OK]")

	return nil
}
