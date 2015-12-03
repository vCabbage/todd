/*
   Hashing functions

   Copyright 2015 - Matt Oswalt
*/

package common

import (
    "crypto/sha256"
    "encoding/hex"
    "fmt"
    "io/ioutil"
)

func GetFileSHA256(filename string) string {
    hasher := sha256.New()
    s, err := ioutil.ReadFile(filename)
    hasher.Write(s)
    FailOnError(err, fmt.Sprintf("Error generating hash for file %s", filename))
    return hex.EncodeToString(hasher.Sum(nil))
}
