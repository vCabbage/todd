/*
    Host Resources - crypto functions

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package hostresources

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"

	"github.com/pkg/errors"
)

func GetFileSHA256(filename string) (string, error) {
	hasher := sha256.New()
	file, err := os.Open(filename)
	if err != nil {
		return "", errors.Wrap(err, "opening file")
	}
	defer file.Close()

	_, err = io.Copy(hasher, file)
	if err != nil {
		return "", errors.Wrap(err, "writing to hasher")
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}
