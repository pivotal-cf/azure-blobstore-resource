package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/christianang/azure-blobstore-resource/api"
	"github.com/christianang/azure-blobstore-resource/azure"
)

func main() {
	var checkRequest api.InRequest
	err := json.NewDecoder(os.Stdin).Decode(&checkRequest)
	if err != nil {
		log.Fatal("failed to decode: ", err)
	}

	azureClient := azure.NewClient(
		checkRequest.Source.StorageAccountName,
		checkRequest.Source.StorageAccountKey,
		checkRequest.Source.Container,
	)
	check := api.NewCheck(azureClient)

	versions, err := check.LatestVersion(checkRequest.Source.VersionedFile)
	if err != nil {
		log.Fatal("failed to get latest version: ", err)
	}

	versionsJSON, err := json.Marshal([]api.Version{versions})
	if err != nil {
		log.Fatal("failed to marshal versions: ", err)
	}

	fmt.Println(string(versionsJSON))
}
