/*
    ToDD UUID generation

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package hostresources

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"strconv"
)

// TruncateID returns a shorthand version of a string identifier for convenience.
// A collision with other shorthands is very unlikely, but possible.
// In case of a collision a lookup with TruncIndex.Get() will fail, and the caller
// will need to use a langer prefix, or the full-length Id.
func TruncateID(id string) string {
	if len(id) < 12 {
		return id
	}
	return id[:12]
}

// GenerateUUID returns an unique id
func GenerateUUID() (string, error) {
	id := make([]byte, 32)
	for {
		if _, err := io.ReadFull(rand.Reader, id); err != nil {
			return "", err
		}
		value := hex.EncodeToString(id)
		// if we try to parse the truncated for as an int and we don't have
		// an error then the value is all numberic and causes issues when
		// used as a hostname. ref #3869
		if _, err := strconv.ParseInt(TruncateID(value), 10, 64); err != nil {
			return value, nil
		}
	}
}
