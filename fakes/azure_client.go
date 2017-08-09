package fakes

import "github.com/Azure/azure-sdk-for-go/storage"

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
}

func (a *AzureClient) ListBlobs(params storage.ListBlobsParameters) (storage.BlobListResponse, error) {
	a.ListBlobsCall.CallCount++
	a.ListBlobsCall.Receives.ListBlobsParameters = params
	return a.ListBlobsCall.Returns.BlobListResponse, a.ListBlobsCall.Returns.Error
}
