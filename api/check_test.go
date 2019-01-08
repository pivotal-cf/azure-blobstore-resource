package api_test

import (
	"errors"
	"time"

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
			var (
				expectedSnapshot time.Time
			)

			BeforeEach(func() {
				expectedSnapshot = time.Date(2017, time.January, 02, 01, 01, 01, 01, time.UTC)
				azureClient.ListBlobsCall.Returns.BlobListResponse = storage.BlobListResponse{
					Blobs: []storage.Blob{
						storage.Blob{
							Name:     "example.json",
							Snapshot: time.Date(2017, time.January, 01, 01, 01, 01, 01, time.UTC),
						},
						storage.Blob{
							Name:     "example.json",
							Snapshot: expectedSnapshot,
						},
						storage.Blob{
							Name:     "example.json",
							Snapshot: time.Date(0001, time.January, 01, 0, 0, 0, 0, time.UTC),
						},
					},
				}
			})

			It("returns just the latest blob snapshot version", func() {
				latestVersion, err := check.LatestVersion("example.json")
				Expect(err).NotTo(HaveOccurred())

				Expect(azureClient.ListBlobsCall.CallCount).To(Equal(1))
				Expect(azureClient.ListBlobsCall.Receives.ListBlobsParameters).To(Equal(storage.ListBlobsParameters{
					Prefix: "example.json",
					Include: &storage.IncludeBlobDataset{
						Snapshots: true,
					},
				}))
				Expect(latestVersion.Snapshot).To(Equal(expectedSnapshot))
			})

			Context("when an error occurs", func() {
				Context("when the azure client fails to list blobs", func() {
					BeforeEach(func() {
						expectedSnapshot = time.Date(2017, time.January, 02, 01, 01, 01, 01, time.UTC)
						azureClient.ListBlobsCall.Returns.Error = errors.New("failed to list blobs")
					})

					It("returns an error", func() {
						_, err := check.LatestVersion("example.json")
						Expect(err).To(MatchError("failed to list blobs"))
					})
				})
			})

			Context("when the file is not found", func() {
				It("returns an error", func() {
					_, err := check.LatestVersion("non-existant.json")
					Expect(err).To(MatchError("failed to find blob: non-existant.json"))
				})
			})
		})
	})
})
