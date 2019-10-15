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
	UploadFromStream(blobName string, blockSize int, stream io.Reader) error
	DownloadBlobToFile(blobName string, file *os.File, blockSize int64, snapshop *time.Time) error
	CreateSnapshot(blobName string) (time.Time, error)
	GetBlobURL(blobName string) (string, error)
}
