package api

import "github.com/toddproject/todd/server/objects"

// TestRunInfo represents the body of the request acceptaed by ServerAPI.Run.
type TestRunInfo struct {
	Name string `json:"name"`
	objects.SourceOverrides
}

// DeleteInfo represents the body of the request accepted by ServerAPI.DeleteObject.
type DeleteInfo struct {
	Label string `json:"label"`
	Type  string `json:"type"`
}
