package azure

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
)

const (
	ChunkSize = 4000000 // 4Mb
)

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

func (c Client) GetRange(blobName string, startRangeInBytes, endRangeInBytes uint64, snapshot time.Time) (io.ReadCloser, error) {
	client, err := storage.NewClient(c.storageAccountName, c.storageAccountKey, c.baseURL, storage.DefaultAPIVersion, true)
	if err != nil {
		return nil, err
	}

	var snapshotPtr *time.Time
	if !snapshot.IsZero() {
		snapshotPtr = &snapshot
	}

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(c.container)
	blob := cnt.GetBlobReference(blobName)
	blobReader, err := blob.GetRange(&storage.GetBlobRangeOptions{
		Range: &storage.BlobRange{
			Start: startRangeInBytes,
			End:   endRangeInBytes,
		},
		GetBlobOptions: &storage.GetBlobOptions{
			Snapshot: snapshotPtr,
		},
	})
	if err != nil {
		return nil, err
	}

	return blobReader, nil
}

func (c Client) UploadFromStream(blobName string, stream io.Reader) error {
	client, err := storage.NewClient(c.storageAccountName, c.storageAccountKey, c.baseURL, storage.DefaultAPIVersion, true)
	if err != nil {
		return err
	}

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(c.container)
	blob := cnt.GetBlobReference(blobName)

	err = blob.CreateBlockBlob(&storage.PutBlobOptions{})
	if err != nil {
		return err
	}

	buffer := make([]byte, ChunkSize)
	blocks := []storage.Block{}
	i := 0
	for {
		bytesRead, err := stream.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}

			return err
		}

		chunk := buffer[:bytesRead]
		blockID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("BlockID{%07d}", i)))
		err = blob.PutBlock(blockID, chunk, &storage.PutBlockOptions{})
		if err != nil {
			return err
		}

		blocks = append(blocks, storage.Block{
			blockID,
			storage.BlockStatusUncommitted,
		})

		i++
	}

	err = blob.PutBlockList(blocks, &storage.PutBlockListOptions{})
	if err != nil {
		return err
	}

	return nil
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
