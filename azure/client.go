package azure

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/Azure/azure-storage-blob-go/azblob"
)

const (
	ChunkSize          = 4000000 // 4Mb
	SnapshotTimeFormat = "2006-01-02T15:04:05.0000000Z"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . AzureClient
type AzureClient interface {
	ListBlobs(params storage.ListBlobsParameters) (storage.BlobListResponse, error)
	GetBlobSizeInBytes(blobName string, snapshot time.Time) (int64, error)
	Get(blobName string, snapshot time.Time) ([]byte, error)
	DownloadBlobToFile(blobName string, file *os.File, snapshop *time.Time, blockSize int64, retryTryTimeout time.Duration) error
	UploadFromStream(blobName string, stream io.Reader, blockSize int, retryTryTimeout time.Duration) error
	CreateSnapshot(blobName string) (time.Time, error)
	GetBlobURL(blobName string) (string, error)
}

type Client struct {
	baseURL            string
	storageAccountName string
	storageAccountKey  string
	container          string
}

func NewClient(baseURL, storageAccountName, storageAccountKey, container string) Client {
	return Client{
		baseURL:            baseURL,
		storageAccountName: storageAccountName,
		storageAccountKey:  storageAccountKey,
		container:          container,
	}
}

func (c Client) ListBlobs(params storage.ListBlobsParameters) (storage.BlobListResponse, error) {
	client, err := storage.NewClient(c.storageAccountName, c.storageAccountKey, c.baseURL, storage.DefaultAPIVersion, true)
	if err != nil {
		return storage.BlobListResponse{}, err
	}

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(c.container)

	return cnt.ListBlobs(params)
}

func (c Client) GetBlobSizeInBytes(blobName string, snapshot time.Time) (int64, error) {
	client, err := storage.NewClient(c.storageAccountName, c.storageAccountKey, c.baseURL, storage.DefaultAPIVersion, true)
	if err != nil {
		return 0, err
	}

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(c.container)
	blob := cnt.GetBlobReference(blobName)

	exists, err := blob.Exists()
	if err != nil {
		return 0, err
	}

	if !exists {
		return 0, fmt.Errorf("%q doesn't exist", blobName)
	}

	var snapshotPtr *time.Time
	if !snapshot.IsZero() {
		snapshotPtr = &snapshot
	}

	err = blob.GetProperties(&storage.GetBlobPropertiesOptions{
		Snapshot: snapshotPtr,
	})
	if err != nil {
		return 0, err
	}

	return blob.Properties.ContentLength, nil

}

func (c Client) Get(blobName string, snapshot time.Time) ([]byte, error) {
	client, err := storage.NewClient(c.storageAccountName, c.storageAccountKey, c.baseURL, storage.DefaultAPIVersion, true)
	if err != nil {
		return []byte{}, err
	}

	var snapshotPtr *time.Time
	if !snapshot.IsZero() {
		snapshotPtr = &snapshot
	}

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(c.container)
	blob := cnt.GetBlobReference(blobName)
	blobReader, err := blob.Get(&storage.GetBlobOptions{
		Snapshot: snapshotPtr,
	})
	if err != nil {
		return []byte{}, err
	}

	defer blobReader.Close()

	data, err := ioutil.ReadAll(blobReader)
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

// DownloadBlobToFile download specified blobName to specified file
func (c Client) DownloadBlobToFile(blobName string, file *os.File, snapshot *time.Time, blockSize int64, retryTryTimeout time.Duration) error {
	u, err := url.Parse(fmt.Sprintf("https://%s.blob.%s/%s/%s",
		c.storageAccountName, c.baseURL, c.container, blobName))
	if err != nil {
		return err
	}

	credential, err := azblob.NewSharedKeyCredential(c.storageAccountName, c.storageAccountKey)
	if err != nil {
		return err
	}

	blobURL := azblob.NewBlobURL(*u, azblob.NewPipeline(credential, azblob.PipelineOptions{
		Retry: azblob.RetryOptions{
			TryTimeout: retryTryTimeout,
		},
	}))

	if snapshot != nil && !snapshot.Equal(time.Time{}) {
		blobURL = blobURL.WithSnapshot(snapshot.Format(SnapshotTimeFormat))
	}

	ctx := context.Background()

	// todo: investigate use of parallelism in options and also retrying downloading of blocks sounds promising
	// as i've seen large downloads (pas tile) fail 80% of the way..
	return azblob.DownloadBlobToFile(ctx, blobURL, 0, 0, file, azblob.DownloadFromBlobOptions{
		BlockSize: blockSize,
	})
}

// UploadFromStream adapted from https://godoc.org/github.com/Azure/azure-storage-blob-go/azblob#example-UploadStreamToBlockBlob
func (c Client) UploadFromStream(blobName string, stream io.Reader, blockSize int, retryTryTimeout time.Duration) error {
	u, err := url.Parse(fmt.Sprintf("https://%s.blob.%s/%s/%s",
		c.storageAccountName, c.baseURL, c.container, blobName))
	if err != nil {
		return err
	}

	credential, err := azblob.NewSharedKeyCredential(c.storageAccountName, c.storageAccountKey)
	if err != nil {
		return err
	}

	blockBlobURL := azblob.NewBlockBlobURL(*u, azblob.NewPipeline(credential, azblob.PipelineOptions{
		Retry: azblob.RetryOptions{
			TryTimeout: retryTryTimeout,
		},
	}))

	ctx := context.Background()

	_, err = azblob.UploadStreamToBlockBlob(ctx, stream, blockBlobURL,
		azblob.UploadStreamToBlockBlobOptions{BufferSize: blockSize, MaxBuffers: 3})

	return err
}

func (c Client) CreateSnapshot(blobName string) (time.Time, error) {
	client, err := storage.NewClient(c.storageAccountName, c.storageAccountKey, c.baseURL, storage.DefaultAPIVersion, true)
	if err != nil {
		return time.Time{}, err
	}

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(c.container)
	blob := cnt.GetBlobReference(blobName)

	snapshot, err := blob.CreateSnapshot(&storage.SnapshotOptions{})
	if err != nil {
		return time.Time{}, err
	}

	return *snapshot, err
}

func (c Client) GetBlobURL(blobName string) (string, error) {
	client, err := storage.NewClient(c.storageAccountName, c.storageAccountKey, c.baseURL, storage.DefaultAPIVersion, true)
	if err != nil {
		return "", err
	}

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(c.container)
	blob := cnt.GetBlobReference(blobName)
	return blob.GetURL(), nil
}
