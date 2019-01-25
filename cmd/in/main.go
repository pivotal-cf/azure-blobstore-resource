package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
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

	baseURL := storage.DefaultBaseURL
	if inRequest.Source.BaseURL != "" {
		baseURL = inRequest.Source.BaseURL
	}

	azureClient := azure.NewClient(
		baseURL,
		inRequest.Source.StorageAccountName,
		inRequest.Source.StorageAccountKey,
		inRequest.Source.Container,
	)
	in := api.NewIn(azureClient)

	var blobName, path string
	var snapshot time.Time
	if inRequest.Source.VersionedFile != "" {
		blobName = inRequest.Source.VersionedFile
		snapshot = inRequest.Version.Snapshot
	} else if inRequest.Source.Regexp != "" {
		blobName = inRequest.Version.Path
		path = inRequest.Version.Path
	}

	err = in.CopyBlobToDestination(
		destinationDirectory,
		blobName,
		snapshot,
	)
	if err != nil {
		log.Fatal("failed to copy blob: ", err)
	}

	url, err := azureClient.GetBlobURL(blobName)
	if err != nil {
		log.Fatal("failed to get blob url: ", err)
	}

	if inRequest.Source.VersionedFile != "" {
		url, err = api.URLAppendTimeStamp(url, inRequest.Version.Snapshot)
		if err != nil {
			log.Fatal("failed to get blob snapshot url: ", err)
		}
	}

	err = ioutil.WriteFile(filepath.Join(destinationDirectory, "url"), []byte(url), os.ModePerm)
	if err != nil {
		log.Fatal("failed to write blob url to output directory: ", err)
	}

	versionsJSON, err := json.Marshal(api.Response{
		Version: api.ResponseVersion{
			Snapshot: snapshot,
			Path:     path,
		},
		Metadata: []api.ResponseMetadata{
			{
				Name:  "filename",
				Value: blobName,
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
