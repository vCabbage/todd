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
		//"626bfabd9634cbbceecc6140f18127f6deb284fcd2966061041dcc7e92b8c7e9": "{\"portal.office.com\":\"{ \\\"num_redirects\\\":0, \\\"size_header\\\":686, \\\"size_request\\\":81, \\\"size_upload\\\":0, \\\"speed_download\\\":240.000, \\\"speed_upload\\\":0.000, \\\"time_redirect\\\":0.000, \\\"time_starttransfer\\\":0.595, \\\"url_effective\\\":\\\"HTTP://portal.office.com/\\\" }\\n\",\"salesforce.com\":\"{ \\\"http_code\\\":\\\"301\\\", \\\"time_namelookup\\\":0.062, \\\"time_connect\\\":0.169, \\\"time_pretransfer\\\":0.169, \\\"time_starttransfer\\\":0.271, \\\"time_total\\\":0.271, \\\"content_type\\\":\\\"\\\", \\\"num_connects\\\":1, \\\"num_redirects\\\":0, \\\"size_download\\\":0, \\\"size_header\\\":101, \\\"size_request\\\":78, \\\"size_upload\\\":0, \\\"speed_download\\\":0.000, \\\"speed_upload\\\":0.000, \\\"time_redirect\\\":0.000, \\\"time_starttransfer\\\":0.271, \\\"url_effective\\\":\\\"HTTP://salesforce.com/\\\" }\\n\"}",
		"2d756c6cd738cce4a709ba7e8432e49ac4032775559422cbbeb4bb62bfbb587a": "{\"4.2.2.2\":{\"avg_latency_ms\":34.309315,\"packet_loss\":0},\"8.8.8.8\":{\"avg_latency_ms\":33.961178,\"packet_loss\":0}}",
		//"fffff": "{\"portal.office.com\":\"{ \"num_redirects\":0, \"size_header\":6.86, }}",
	}

	_, err := cleanTestData(dirtyData)
	if err != nil {
		t.Fatalf(fmt.Sprint(err))
	}

}
