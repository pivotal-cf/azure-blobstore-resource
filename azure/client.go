package azure

import (
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

func (c Client) Get(blobName string, snapshot time.Time) ([]byte, error) {
	client, err := storage.NewBasicClient(c.storageAccountName, c.storageAccountKey)
	if err != nil {
		return []byte{}, err
	}

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(c.container)
	blob := cnt.GetBlobReference(blobName)
	blobReader, err := blob.Get(&storage.GetBlobOptions{
		Snapshot: &snapshot,
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
