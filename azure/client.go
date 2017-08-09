package azure

import (
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
