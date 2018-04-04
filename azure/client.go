package azure

import (
	"io"
	"io/ioutil"
	"time"
	"github.com/Azure/azure-sdk-for-go/storage"
)

type Client struct {
	storageAccountName string
	storageAccountKey  string
	container          string
}

func NewClient(storageAccountName, storageAccountKey, container string) Client {
	return Client{
		storageAccountName: storageAccountName,
		storageAccountKey:  storageAccountKey,
		container:          container,
	}
}

func (c Client) ListBlobs(params storage.ListBlobsParameters) (storage.BlobListResponse, error) {
	client, err := storage.NewBasicClient(c.storageAccountName, c.storageAccountKey)
	if err != nil {
		return storage.BlobListResponse{}, err
	}

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(c.container)

	return cnt.ListBlobs(params)
}

func (c Client) Get(blobName string) ([]byte, error) {
	client, err := storage.NewBasicClient(c.storageAccountName, c.storageAccountKey)
	if err != nil {
		return []byte{}, err
	}

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(c.container)
	blob := cnt.GetBlobReference(blobName)
	blobReader, err := blob.Get(&storage.GetBlobOptions{})
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

func (c Client) UploadFromStream(blobName string, stream io.Reader) error {
	client, err := storage.NewBasicClient(c.storageAccountName, c.storageAccountKey)
	if err != nil {
		return err
	}

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(c.container)
	blob := cnt.GetBlobReference(blobName)

	err = blob.CreateBlockBlobFromReader(stream, nil)
	if err != nil {
		return err
	}

	return nil
}

func (c Client) CreateSnapshot(blobName string) (time.Time, error) {
	client, err := storage.NewBasicClient(c.storageAccountName, c.storageAccountKey)
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
	client, err := storage.NewBasicClient(c.storageAccountName, c.storageAccountKey)
	if err != nil {
		return "", err
	}

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(c.container)
	blob := cnt.GetBlobReference(blobName)
	return blob.GetURL(), nil
}
