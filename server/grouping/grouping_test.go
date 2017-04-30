/*
   Grouping tests

   Copyright 2017 Matt Oswalt. Use or modification of this
   source code is governed by the license provided here:
   https://github.com/toddproject/todd/blob/master/LICENSE
*/

package grouping

import (
	"testing"
)

func TestGrouping(t *testing.T) {

	matchStatements := []map[string]string{
		{"hostname": "toddtestagent1"},
		{"hostname": "toddtestagent2"},
		{"hostname": "toddtestagent3"},
		{"hostname": "toddtestagent4"},
		{"hostname": "toddtestagent5"},
		{"hostname": "toddtestagent6"},
	}

	factmap := map[string][]string{
		"Addresses": {"127.0.0.1", "::1"},
		"Hostname":  {"toddtestagent1"},
	}

	result := isInGroup(matchStatements, factmap)
	if !result {
		t.Fatalf("Not in the group and it should be")
	}
}
