/*
    Tests for uuid functions

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package hostresources

import (
	"regexp"
	"testing"
	"unicode/utf8"
)

var validShortID = regexp.MustCompile("^[a-z0-9]{12}$")

// isShortID determine if an arbitrary string *looks like* a short ID.
func isShortID(id string) bool {
	return validShortID.MatchString(id)
}

// TestGenerateUuid tests that GenerateUuid generates a valid UUID (by length)
func TestGenerateUuid(t *testing.T) {
	thisUUID, err := GenerateUUID()
	if err != nil {
		t.Fatal(err)
	}
	if utf8.RuneCountInString(thisUUID) != 64 {
		t.Fatalf("Invalid UUID generated")
	}
}

// TestTruncateID will test that TruncateID properly truncates a UUID
func TestTruncateID(t *testing.T) {
	thisUUID := TruncateID("eb00b2a4f7e58ea03e05b81839a72ee810250010aab27431edebb64cb73aae27")
	if thisUUID != "eb00b2a4f7e5" {
		t.Fatalf("Invalid UUID truncation")
	}
}

// TestIsShortIDValid will test to ensure that IsShortIDValid is able to accurately identify a valid shortened UUID
func TestIsShortIDValid(t *testing.T) {
	if !isShortID("eb00b2a4f7e5") {
		t.Fatalf("IsShortIDValid is not returning a correct result")
	}
}

// TestIsShortIDValid will test to ensure that IsShortIDValid is able to accurately identify an invalid shortened UUID
func TestIsShortIDInvalid(t *testing.T) {
	if isShortID("eb00bxxxxxxxxx2a4f7e5") {
		t.Fatalf("IsShortIDValid is not returning a correct result")
	}
}
