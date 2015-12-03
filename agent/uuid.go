/*
   ToDD Agent UUID generation

   Copyright 2015 - Matt Oswalt
*/

package main

import (
    "crypto/rand"
    "encoding/hex"
    "io"
    "regexp"
    "strconv"
)

var UUID []byte

// func generateUuid() string {
//  // TODO(moswalt): This is linux-specific. Maybe use a lib?
//  uuid, err := exec.Command("uuidgen").Output()
//  common.FailOnError(err, "Failed to generate a UUID")

//  // Strip newline - also probably not needed with a proper UUID lib
//  //uuid_str := strings.Replace(string(uuid[:]), "\n", "", -1)
//  uuid_str := string(uuid[:])
//  uuid_str = strings.TrimRight(uuid_str, "\n")

//  return uuid_str
// }

var validShortID = regexp.MustCompile("^[a-z0-9]{12}$")

// Determine if an arbitrary string *looks like* a short ID.
func IsShortID(id string) bool {
    return validShortID.MatchString(id)
}

const shortLen = 12

// TruncateID returns a shorthand version of a string identifier for convenience.
// A collision with other shorthands is very unlikely, but possible.
// In case of a collision a lookup with TruncIndex.Get() will fail, and the caller
// will need to use a langer prefix, or the full-length Id.
func TruncateID(id string) string {
    trimTo := shortLen
    if len(id) < shortLen {
        trimTo = len(id)
    }
    return id[:trimTo]
}

// generateUuid returns an unique id
func generateUuid() string {
    for {
        id := make([]byte, 32)
        if _, err := io.ReadFull(rand.Reader, id); err != nil {
            panic(err) // This shouldn't happen
        }
        value := hex.EncodeToString(id)
        // if we try to parse the truncated for as an int and we don't have
        // an error then the value is all numberic and causes issues when
        // used as a hostname. ref #3869
        if _, err := strconv.ParseInt(TruncateID(value), 10, 64); err == nil {
            continue
        }
        return value
    }
}
