/*
	ToDD task - download collectors

	Copyright 2016 Matt Oswalt. Use or modification of this
	source code is governed by the license provided here:
	https://github.com/toddproject/todd/blob/master/LICENSE
*/

package tasks

import (
	"errors"
	"fmt"
	"net/http"
	"strings"

	log "github.com/Sirupsen/logrus"
)

// DownloadAssetTask defines this particular task. It contains definitions not only for the task message, but
// also HTTPClient, filesystem, and i/o system abstractions to be more conducive to unit testing.
type DownloadAssetTask struct {
	BaseTask
	HTTPClient   *http.Client `json:"-"`
	Fs           FileSystem   `json:"-"`
	Ios          IoSystem     `json:"-"`
	CollectorDir string       `json:"-"`
	Assets       []string     `json:"assets"`
	TestletDir   string       `json:"-"`
}

// Run contains the logic necessary to perform this task on the agent. This particular task will download all required assets,
// copy them into the appropriate directory, and ensure that the execute permission is given to each collector file.
func (dat DownloadAssetTask) Run() error {

	// Iterate over the slice of collectors and download them.
	for x := range dat.Assets {

		assetURL := dat.Assets[x]

		assetDir := ""

		switch {
		case strings.Contains(assetURL, "factcollectors"):
			assetDir = dat.CollectorDir
		case strings.Contains(assetURL, "testlets"):
			assetDir = dat.TestletDir
		default:
			errorMsg := "Invalid asset download URL received"
			log.Error(errorMsg)
			return errors.New(errorMsg)
		}

		err := dat.downloadAsset(assetURL, assetDir)
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return nil
}

// downloadAsset will download an asset at the specified URL, into the specified directory
func (dat DownloadAssetTask) downloadAsset(url, directory string) error {

	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	fileName = fmt.Sprintf("%s/%s", directory, fileName)
	log.Info("Downloading ", url, " to ", fileName)

	// TODO: What if this already exists? Consider checking file existence first with io.IsExist?
	output, err := dat.Fs.Create(fileName)
	if err != nil {
		return fmt.Errorf("Error while creating %s - %v", fileName, err)
	}
	defer output.Close()

	response, err := dat.HTTPClient.Get(url)
	if err != nil || response.StatusCode != 200 {
		// If we have a problem retrieving the testlet, we want to return immediately,
		// instead of writing an empty file to disk
		log.Errorf("Error while downloading '%s': %s", url, response.Status)
		return err
	}
	defer response.Body.Close()

	n, err := dat.Ios.Copy(output, response.Body)
	if err != nil {
		log.Error(fmt.Sprintf("Error while writing '%s' to disk", url))
		return err
	}
	err = dat.Fs.Chmod(fileName, 0744)
	if err != nil {

		// TODO(mierdin): currently unhandled; may want to do something further
		log.Warn("Problem setting execute permission on downloaded script")
	}

	log.Info(n, " bytes downloaded.")
	return nil
}
