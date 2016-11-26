/*
    ToDD Client API Calls for "todd groups"

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
	"text/tabwriter"

	"github.com/toddproject/todd/hostresources"
)

// Groups will query ToDD for a map containing current agent-to-group mappings
func (c *ClientAPI) Groups() error {
	url := c.baseURL + "/groups"
	resp, err := c.http.Get(url)
	if err != nil {
		return err
	}
	defer io.Copy(ioutil.Discard, resp.Body)
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	// Read the content into a byte array
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	// Marshal API data into map
	var groupmap map[string]string
	err = json.Unmarshal(body, &groupmap)
	if err != nil {
		return err
	}

	// Format in tab-separated columns with a tab stop of 8.
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintln(w, "UUID\tGROUP NAME")

	for agentUUID, groupName := range groupmap {
		fmt.Fprintf(
			w,
			"%s\t%s\n",
			hostresources.TruncateID(agentUUID),
			groupName,
		)
	}
	fmt.Fprintln(w)
	w.Flush()

	return nil
}
