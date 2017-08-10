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
	sourceDirectory := os.Args[1]

	var outRequest api.OutRequest
	err := json.NewDecoder(os.Stdin).Decode(&outRequest)
	if err != nil {
		log.Fatal("failed to decode: ", err)
	}

	azureClient := azure.NewClient(
		outRequest.Source.StorageAccountName,
		outRequest.Source.StorageAccountKey,
		outRequest.Source.Container,
	)
	out := api.NewOut(azureClient)

	snapshot, err := out.UploadFileToBlobstore(
		sourceDirectory,
		outRequest.Params.File,
		outRequest.Source.VersionedFile,
	)
	if err != nil {
		log.Fatal("failed to upload blob: ", err)
	}

	versionsJSON, err := json.Marshal(api.Response{
		Version: api.ResponseVersion{
			Snapshot: snapshot,
		},
	})
	if err != nil {
		log.Fatal("failed to marshal output: ", err)
	}

	fmt.Println(string(versionsJSON))
}
