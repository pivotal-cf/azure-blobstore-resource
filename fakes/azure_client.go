package fakes

import (
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
)

type AzureClient struct {
	ListBlobsCall struct {
		CallCount int
		Receives  struct {
			ListBlobsParameters storage.ListBlobsParameters
		}
		Returns struct {
			BlobListResponse storage.BlobListResponse
			Error            error
		}
	}
	GetCall struct {
		CallCount int
		Receives  struct {
			BlobName string
			Snapshot time.Time
		}
		Returns struct {
			BlobData []byte
			Error    error
		}
	}
}

func (a *AzureClient) ListBlobs(params storage.ListBlobsParameters) (storage.BlobListResponse, error) {
	a.ListBlobsCall.CallCount++
	a.ListBlobsCall.Receives.ListBlobsParameters = params
	return a.ListBlobsCall.Returns.BlobListResponse, a.ListBlobsCall.Returns.Error
}

func (a *AzureClient) Get(blobName string, snapshot time.Time) ([]byte, error) {
	a.GetCall.CallCount++
	a.GetCall.Receives.BlobName = blobName
	a.GetCall.Receives.Snapshot = snapshot
	return a.GetCall.Returns.BlobData, a.GetCall.Returns.Error
}
