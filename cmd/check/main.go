package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/pivotal-cf/azure-blobstore-resource/api"
	"github.com/pivotal-cf/azure-blobstore-resource/azure"
)

func main() {
	var checkRequest api.InRequest
	err := json.NewDecoder(os.Stdin).Decode(&checkRequest)
	if err != nil {
		log.Fatal("failed to decode: ", err)
	}

	baseURL := storage.DefaultBaseURL
	if checkRequest.Source.BaseURL != "" {
		baseURL = checkRequest.Source.BaseURL
	}

	azureClient := azure.NewClient(
		baseURL,
		checkRequest.Source.StorageAccountName,
		checkRequest.Source.StorageAccountKey,
		checkRequest.Source.Container,
	)
	check := api.NewCheck(azureClient)

	var versions []api.Version
	if checkRequest.Source.VersionedFile != "" {
		versions, err = check.VersionsSince(checkRequest.Source.VersionedFile, checkRequest.Version.Snapshot)
		if err != nil {
			log.Fatal("failed to get latest version: ", err)
		}
	} else if checkRequest.Source.Regexp != "" {
		version, err := check.LatestVersionRegexp(checkRequest.Source.Regexp)
		if err != nil {
			log.Fatal("failed to get latest version from regexp: ", err)
		}
		versions = []api.Version{version}
	} else {
		log.Fatal("must supply either versioned_file or regexp in source parameters", err)
	}

	versionsJSON, err := json.Marshal(versions)
	if err != nil {
		log.Fatal("failed to marshal versions: ", err)
	}

	fmt.Println(string(versionsJSON))
}
