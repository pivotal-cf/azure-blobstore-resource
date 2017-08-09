package api

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
)

type azureClient interface {
	ListBlobs(params storage.ListBlobsParameters) (storage.BlobListResponse, error)
}

type Version struct {
	Snapshot time.Time `json:"snapshot"`
}

type Check struct {
	azureClient azureClient
}

func NewCheck(azureClient azureClient) Check {
	return Check{azureClient: azureClient}
}

func (c Check) LatestVersion(filename string) (Version, error) {
	blobListResponse, err := c.azureClient.ListBlobs(storage.ListBlobsParameters{
		Prefix: filename,
		Include: &storage.IncludeBlobDataset{
			Snapshots: true,
		},
	})
	if err != nil {
		return Version{}, err
	}

	var latestSnapshot time.Time
	for _, blob := range blobListResponse.Blobs {
		if blob.Name == filename {
			if blob.Snapshot.After(latestSnapshot) {
				latestSnapshot = blob.Snapshot
			}
		}
	}

	return Version{
		Snapshot: latestSnapshot,
	}, nil
}
