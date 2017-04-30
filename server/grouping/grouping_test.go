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

// TestHostnameRegex tests hostname statement with regex
func TestHostnameRegex(t *testing.T) {

	matchStatements := []map[string]string{
		{"hostname": "toddtestagent[1-6]"},
	}

	result := isInGroup(matchStatements, map[string][]string{
		"Addresses": {"127.0.0.1", "::1"},
		"Hostname":  {"toddtestagent7"},
	})
	if result {
		t.Fatalf("In the group and it should not be")
	}

	result = isInGroup(matchStatements, map[string][]string{
		"Addresses": {"127.0.0.1", "::1"},
		"Hostname":  {"toddtestagent6"},
	})
	if !result {
		t.Fatalf("Not in the group and it should be")
	}
}

// TestMultipleHostnames tests with several hostname statements instead of a single regex
func TestMultipleHostnames(t *testing.T) {

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
		"Hostname":  {"toddtestagent6"},
	}

	result := isInGroup(matchStatements, factmap)
	if !result {
		t.Fatalf("Not in the group and it should be")
	}
}

// TestInSubnet tests within_subnet statement
func TestInSubnet(t *testing.T) {

	matchStatements := []map[string]string{
		{"within_subnet": "192.168.0.0/24"},
	}

	result := isInGroup(matchStatements, map[string][]string{
		"Addresses": {"127.0.0.1", "::1"},
		"Hostname":  {"todddev"},
	})
	if result {
		t.Fatalf("In the group and it should not be")
	}

	result = isInGroup(matchStatements, map[string][]string{
		"Addresses": {"192.168.0.1", "::1"},
		"Hostname":  {"todddev"},
	})
	if !result {
		t.Fatalf("Not in the group and it should be")
	}

}
