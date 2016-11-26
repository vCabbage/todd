/*
   ToDD Client API Calls

    Copyright 2016 Matt Oswalt. Use or modification of this
    source code is governed by the license provided here:
    https://github.com/toddproject/todd/blob/master/LICENSE
*/

package api

import (
	"fmt"
	"net/http"
	"time"
)

type ClientAPI struct {
	http    *http.Client
	host    string
	baseURL string
}

func New(host string, port int) *ClientAPI {
	return &ClientAPI{
		http: &http.Client{
			Timeout: 5 * time.Second, // TODO: arbitrary value, might make sense to reduce
		},
		host:    host,
		baseURL: fmt.Sprintf("http://%s:%d/v1", host, port),
	}
}
