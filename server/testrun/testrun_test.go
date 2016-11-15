/*
   Tests for "testrun" package

   Copyright 2016 Matt Oswalt. Use or modification of this
   source code is governed by the license provided here:
   https://github.com/toddproject/todd/blob/master/LICENSE
*/

package testrun

import (
	"fmt"
	"testing"
)

func TestCleanData(t *testing.T) {

	// dirtyData is a rough example of what would be passed in to the "cleanTestData" function.
	// It's mock data, but it uses a variety of datatypes like strings, ints, and floats, which
	// is an important thing to test
	dirtyData := map[string]string{
		"2d756c6cd738cce4a709ba7e8432e49ac4032775559422cbbeb4bb62bfbb587a": "{\"4.2.2.2\":{\"avg_latency_ms\":34.309315,\"packet_loss\":0},\"8.8.8.8\":{\"avg_latency_ms\":33.961178,\"packet_loss\":0}}",
	}

	_, err := cleanTestData(dirtyData)
	if err != nil {
		t.Fatalf(fmt.Sprint(err))
	}

}
