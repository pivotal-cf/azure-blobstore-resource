package api

import (
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
)

type azureClient interface {
	ListBlobs(params storage.ListBlobsParameters) (storage.BlobListResponse, error)
	Get(blobName string) ([]byte, error)
	UploadFromStream(blobName string, stream io.Reader) error
	CreateSnapshot(blobName string) (time.Time, error)
}
