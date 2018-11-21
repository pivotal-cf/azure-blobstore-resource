package api

import (
	"io"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
)

type azureClient interface {
	ListBlobs(params storage.ListBlobsParameters) (storage.BlobListResponse, error)
	Get(blobName string, snapshot time.Time) ([]byte, error)
	GetRange(blobName string, startRangeInBytes, endRangeInBytes uint64, snapshot time.Time) (io.ReadCloser, error)
	GetBlobSizeInBytes(blobName string, snapshop time.Time) (int64, error)
	UploadFromStream(blobName string, stream io.Reader) error
	CreateSnapshot(blobName string) (time.Time, error)
	GetBlobURL(blobName string) (string, error)
}
