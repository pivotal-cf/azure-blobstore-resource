package api

import (
	"io"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
)

type azureClient interface {
	ListBlobs(params storage.ListBlobsParameters) (storage.BlobListResponse, error)
	Get(blobName string, snapshot time.Time) ([]byte, error)
	GetBlobSizeInBytes(blobName string, snapshop time.Time) (int64, error)
	UploadFromStream(blobName string, stream io.Reader, blockSize int, retryTryTimeout time.Duration) error
	DownloadBlobToFile(blobName string, file *os.File, snapshop *time.Time, blockSize int64, retryTryTimeout time.Duration) error
	CreateSnapshot(blobName string) (time.Time, error)
	GetBlobURL(blobName string) (string, error)
}
