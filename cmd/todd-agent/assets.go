/*
	ToDD Agent Asset management

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package main

import (
	"os"
	"path/filepath"

	log "github.com/Sirupsen/logrus"

	"github.com/toddproject/todd/hostresources"
)

// getLocalAssets gathers all currently installed assets, and generates a map of their names and hashes.
func getLocalAssets(dir string) (factcollectors, testlets map[string]string, err error) {
	factcollectors, err = getAssets(dir, "factcollectors")
	if err != nil {
		return nil, nil, err
	}
	testlets, err = getAssets(dir, "testlets")
	return factcollectors, testlets, err

}

func getAssets(optDir, typ string) (map[string]string, error) {
	foundAssets := make(map[string]string)

	// set up a directory for this particular asset type
	assetDir := filepath.Join(optDir, "assets", typ)

	// create fact collector directory if needed
	err := os.MkdirAll(assetDir, 0777)
	if err != nil {
		log.Error("Problem creating asset directory")
		os.Exit(1)
	}

	// this is the function that will generate a hash for a file and add it to our asset map
	discoverAssets := func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		// Generate hash
		foundAssets[f.Name()] = hostresources.GetFileSHA256(path)
		log.Debugf("Asset found locally: %s (with hash %s)", f.Name(), hostresources.GetFileSHA256(path))
		return nil
	}

	// Perform above Walk function (discover_assets) on the collector directory
	err = filepath.Walk(assetDir, discoverAssets)
	return foundAssets, err
}
