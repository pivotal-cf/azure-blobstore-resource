package api_test

import (
	"errors"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/pivotal-cf/azure-blobstore-resource/api"
	"github.com/pivotal-cf/azure-blobstore-resource/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check", func() {
	var (
		azureClient *fakes.AzureClient
		check       api.Check
	)

	BeforeEach(func() {
		azureClient = &fakes.AzureClient{}
		check = api.NewCheck(azureClient)
	})

	Describe("LatestVersion", func() {
		Context("when given a filename", func() {

			BeforeEach(func() {
				azureClient.ListBlobsCall.Returns.BlobListResponse = storage.BlobListResponse{
					Blobs: []storage.Blob{
						storage.Blob{
							Name:     "example.json",
						},
						storage.Blob{
							Name:     "example.json",
						},
						storage.Blob{
							Name:     "example.json",
						},
					},
				}
			})

			Context("when an error occurs", func() {
				Context("when the azure client fails to list blobs", func() {
					BeforeEach(func() {
						azureClient.ListBlobsCall.Returns.Error = errors.New("failed to list blobs")
					})

					It("returns an error", func() {
						_, err := check.LatestVersion("example.json")
						Expect(err).To(MatchError("failed to list blobs"))
					})
				})
			})
		})
	})
})
