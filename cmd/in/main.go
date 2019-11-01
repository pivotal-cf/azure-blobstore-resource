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
	"github.com/Azure/azure-storage-blob-go/azblob"
	"github.com/pivotal-cf/azure-blobstore-resource/api"
	"github.com/pivotal-cf/azure-blobstore-resource/azure"
)

const (
	DefaultRetryTryTimeout = time.Duration(0)
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

	var blobName, versionPath string
	var snapshot *time.Time
	if inRequest.Source.VersionedFile != "" {
		blobName = inRequest.Source.VersionedFile
		snapshot = &inRequest.Version.Snapshot
	} else if inRequest.Source.Regexp != "" {
		blobName = inRequest.Version.Path
		versionPath = inRequest.Version.Path
	}

	blockSize := azblob.BlobDefaultDownloadBlockSize
	if inRequest.Params.BlockSize != nil {
		blockSize = *inRequest.Params.BlockSize
	}

	retryTryTimeout := DefaultRetryTryTimeout
	if inRequest.Params.Retry.TryTimeout != nil {
		retryTryTimeout = time.Duration(*inRequest.Params.Retry.TryTimeout)
	}

	if !inRequest.Params.SkipDownload {
		err = in.CopyBlobToDestination(
			destinationDirectory,
			blobName,
			snapshot,
			blockSize,
			retryTryTimeout,
		)
		if err != nil {
			log.Fatal("failed to copy blob: ", err)
		}

		if inRequest.Params.Unpack {
			err = in.UnpackBlob(filepath.Join(destinationDirectory, blobName))
			if err != nil {
				log.Fatal("failed to unpack blob: ", err)
			}
		}
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

	err = ioutil.WriteFile(filepath.Join(destinationDirectory, "version"),
		[]byte(inRequest.Version.Version), os.ModePerm)
	if err != nil {
		log.Fatal("failed to write blob version to output directory: ", err)
	}

	versionsJSON, err := json.Marshal(api.Response{
		Version: api.ResponseVersion{
			Snapshot: snapshot,
			Path:     versionPath,
			Version:  inRequest.Version.Version,
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
