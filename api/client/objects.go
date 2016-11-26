/*
    ToDD Client API Calls for "todd objects"

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"text/tabwriter"

	"github.com/toddproject/todd/server/objects"
)

// Objects will query ToDD for all objects, with the type requested in the sub-arguments, and then display a list of those
// objects to the user.
func (c *ClientAPI) Objects(objType string) error {
	// If no subarg was provided, instruct the user to provide the object type
	if objType == "" {
		return errors.New("Please provide the object type")
	}

	url := fmt.Sprintf("%s/object/%s", c.baseURL, objType)

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

	parsedObjects := objects.ParseToddObjects(body)

	// Format in tab-separated columns with a tab stop of 8.
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 0, '\t', 0)
	fmt.Fprintln(w, "LABEL\tTYPE\tSPEC\t")
	for i := range parsedObjects {

		fmt.Fprintf(
			w,
			"%s\t%s\t%s\n",
			parsedObjects[i].GetLabel(),
			parsedObjects[i].GetType(),
			parsedObjects[i].GetSpec(),
		)
	}
	fmt.Fprintln(w)
	w.Flush()

	return nil
}
