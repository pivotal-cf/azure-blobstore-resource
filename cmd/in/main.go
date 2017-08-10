package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/christianang/azure-blobstore-resource/api"
	"github.com/christianang/azure-blobstore-resource/azure"
)

type Output struct {
	Version OutputVersion `json:"version"`
}

type OutputVersion struct {
	Snapshot time.Time `json:"snapshot"`
}

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

	versionsJSON, err := json.Marshal(Output{
		Version: OutputVersion{
			Snapshot: inRequest.Version.Snapshot,
		},
	})
	if err != nil {
		log.Fatal("failed to marshal versions: ", err)
	}

	fmt.Println(string(versionsJSON))
}
