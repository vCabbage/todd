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
	// It's obviously mock data, but it uses a variety of datatypes like strings, ints, and floats, which
	// is an important thing to test
	dirtyData := map[string]string{
		"626bfabd9634cbbceecc6140f18127f6deb284fcd2966061041dcc7e92b8c7e9": "{\"portal.office.com\":\"{ \\\"num_redirects\\\":0, \\\"size_download\\\":, \\\"size_header\\\":686, \\\"size_request\\\":81, \\\"size_upload\\\":0, \\\"speed_download\\\":240.000, \\\"speed_upload\\\":0.000, \\\"time_redirect\\\":0.000, \\\"time_starttransfer\\\":0.595, \\\"url_effective\\\":\\\"HTTP://portal.office.com/\\\" }\\n\",\"salesforce.com\":\"{ \\\"http_code\\\":\\\"301\\\", \\\"time_namelookup\\\":0.062, \\\"time_connect\\\":0.169, \\\"time_pretransfer\\\":0.169, \\\"time_starttransfer\\\":0.271, \\\"time_total\\\":0.271, \\\"content_type\\\":\\\"\\\", \\\"http_connect\\\":000, \\\"num_connects\\\":1, \\\"num_redirects\\\":0, \\\"size_download\\\":0, \\\"size_header\\\":101, \\\"size_request\\\":78, \\\"size_upload\\\":0, \\\"speed_download\\\":0.000, \\\"speed_upload\\\":0.000, \\\"time_redirect\\\":0.000, \\\"time_starttransfer\\\":0.271, \\\"url_effective\\\":\\\"HTTP://salesforce.com/\\\" }\\n\"}",
		// "983f4d47908175bbc88ffe7df31644a88a6843a78350be9ee16dfc280af8c349": "{\"portal.office.com\":\"{ \\\"http_code\\\":\\\"302\\\", \\\"time_namelookup\\\":0.064, \\\"time_connect\\\":0.111, \\\"time_pretransfer\\\":0.111, \\\"time_starttransfer\\\":0.150, \\\"time_total\\\":0.150, \\\"content_type\\\":\\\"\\\", \\\"http_connect\\\":000, \\\"num_connects\\\":1, \\\"num_redirects\\\":0, \\\"size_download\\\":, \\\"size_header\\\":686, \\\"size_request\\\":81, \\\"size_upload\\\":0, \\\"speed_download\\\":950.000, \\\"speed_upload\\\":0.000, \\\"time_redirect\\\":0.000, \\\"time_starttransfer\\\":0.150, \\\"url_effective\\\":\\\"HTTP://portal.office.com/\\\" }\\n\",\"salesforce.com\":\"{ \\\"http_code\\\":\\\"301\\\", \\\"time_namelookup\\\":0.062, \\\"time_connect\\\":0.172, \\\"time_pretransfer\\\":0.173, \\\"time_starttransfer\\\":0.275, \\\"time_total\\\":0.275, \\\"content_type\\\":\\\"\\\", \\\"http_connect\\\":000, \\\"num_connects\\\":1, \\\"num_redirects\\\":0, \\\"size_download\\\":0, \\\"size_header\\\":101, \\\"size_request\\\":78, \\\"size_upload\\\":0, \\\"speed_download\\\":0.000, \\\"speed_upload\\\":0.000, \\\"time_redirect\\\":0.000, \\\"time_starttransfer\\\":0.275, \\\"url_effective\\\":\\\"HTTP://salesforce.com/\\\" }\\n\"}",
		// "f2bebacce45fe9cde611deb8bbf9d5b65c8d15cfbe323d42c10e714b5f0b9fbd": "{\"portal.office.com\":\"{ \\\"http_code\\\":\\\"302\\\", \\\"time_namelookup\\\":0.062, \\\"time_connect\\\":0.113, \\\"time_pretransfer\\\":0.113, \\\"time_starttransfer\\\":0.154, \\\"time_total\\\":0.154, \\\"content_type\\\":\\\"\\\", \\\"http_connect\\\":000, \\\"num_connects\\\":1, \\\"num_redirects\\\":0, \\\"size_download\\\":, \\\"size_header\\\":686, \\\"size_request\\\":81, \\\"size_upload\\\":0, \\\"speed_download\\\":929.000, \\\"speed_upload\\\":0.000, \\\"time_redirect\\\":0.000, \\\"time_starttransfer\\\":0.154, \\\"url_effective\\\":\\\"HTTP://portal.office.com/\\\" }\\n\",\"salesforce.com\":\"{ \\\"http_code\\\":\\\"301\\\", \\\"time_namelookup\\\":0.061, \\\"time_connect\\\":0.182, \\\"time_pretransfer\\\":0.182, \\\"time_starttransfer\\\":0.288, \\\"time_total\\\":0.288, \\\"content_type\\\":\\\"\\\", \\\"http_connect\\\":000, \\\"num_connects\\\":1, \\\"num_redirects\\\":0, \\\"size_download\\\":0, \\\"size_header\\\":101, \\\"size_request\\\":78, \\\"size_upload\\\":0, \\\"speed_download\\\":0.000, \\\"speed_upload\\\":0.000, \\\"time_redirect\\\":0.000, \\\"time_starttransfer\\\":0.288, \\\"url_effective\\\":\\\"HTTP://salesforce.com/\\\" }\\n\"}",
		// "uuid12345": "{\"portal.office.com\":\"{ \\\"http_code\\\":\\\"poopdeooo\\\", \\\"time_namelookup\\\":0.061, \\\"time_connect\\\":111 }\"}",
	}

	_, err := cleanTestData(dirtyData)
	if err != nil {
		t.Fatalf(fmt.Sprint(err))
	}

}
