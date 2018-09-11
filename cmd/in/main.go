package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/pivotal-cf/azure-blobstore-resource/api"
	"github.com/pivotal-cf/azure-blobstore-resource/azure"
)

func main() {
	destinationDirectory := os.Args[1]

	var inRequest api.InRequest
	err := json.NewDecoder(os.Stdin).Decode(&inRequest)
	if err != nil {
		log.Fatal("failed to decode: ", err)
	}

	azureClient := azure.NewClient(
		inRequest.Source.StorageAccountName,
		inRequest.Source.StorageAccountKey,
		inRequest.Source.Container,
	)
	in := api.NewIn(azureClient)

	err = in.CopyBlobToDestination(
		destinationDirectory,
		inRequest.Source.VersionedFile,
		inRequest.Version.Snapshot,
	)
	if err != nil {
		log.Fatal("failed to copy blob: ", err)
	}

	url, err := azureClient.GetBlobURL(inRequest.Source.VersionedFile, inRequest.Version.Snapshot)
	if err != nil {
		log.Fatal("failed to get blob url: ", err)
	}

	versionsJSON, err := json.Marshal(api.Response{
		Version: api.ResponseVersion{
			Snapshot: inRequest.Version.Snapshot,
		},
		Metadata: []api.ResponseMetadata{
			{
				Name:  "filename",
				Value: inRequest.Source.VersionedFile,
			},
			{
				Name:  "url",
				Value: url,
			},
		},
	})
	if err != nil {
		log.Fatal("failed to marshal output: ", err)
	}

	fmt.Println(string(versionsJSON))
}
