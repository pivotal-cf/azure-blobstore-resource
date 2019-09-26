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

	Describe("VersionsSince", func() {
		Context("given a filename", func() {
			var (
				expectedSnapshotCurrent time.Time
				expectedSnapshotNew     time.Time
				expectedSnapshotNewer   time.Time
			)

			BeforeEach(func() {
				expectedSnapshotCurrent = time.Date(2017, time.January, 02, 01, 01, 01, 01, time.UTC)
				expectedSnapshotNew = time.Date(2017, time.January, 03, 01, 01, 01, 01, time.UTC)
				expectedSnapshotNewer = time.Date(2017, time.January, 04, 01, 01, 01, 01, time.UTC)
				azureClient.ListBlobsCall.Returns.BlobListResponse = storage.BlobListResponse{
					Blobs: []storage.Blob{
						storage.Blob{
							Name:     "example.json",
							Snapshot: time.Date(2017, time.January, 01, 01, 01, 01, 01, time.UTC),
						},
						storage.Blob{
							Name:     "example.json",
							Snapshot: expectedSnapshotNewer,
						},
						storage.Blob{
							Name:     "example.json",
							Snapshot: expectedSnapshotCurrent,
						},
						storage.Blob{
							Name:     "example.json",
							Snapshot: expectedSnapshotNew,
						},
						storage.Blob{
							Name:     "example.json",
							Snapshot: time.Date(0001, time.January, 01, 0, 0, 0, 0, time.UTC),
						},
					},
				}
			})

			It("returns versions from current to latest for blob", func() {
				latestVersions, err := check.VersionsSince("example.json", expectedSnapshotCurrent)
				Expect(err).NotTo(HaveOccurred())

				Expect(azureClient.ListBlobsCall.CallCount).To(Equal(1))
				Expect(azureClient.ListBlobsCall.Receives.ListBlobsParameters).To(Equal(storage.ListBlobsParameters{
					Prefix: "example.json",
					Include: &storage.IncludeBlobDataset{
						Snapshots: true,
						Copy:      true,
					},
				}))
				Expect(latestVersions).To(HaveLen(3))
				Expect(latestVersions[0].Snapshot).To(Equal(&expectedSnapshotCurrent))
				Expect(latestVersions[1].Snapshot).To(Equal(&expectedSnapshotNew))
				Expect(latestVersions[2].Snapshot).To(Equal(&expectedSnapshotNewer))
			})

			Context("when an error occurs", func() {
				Context("when the azure client fails to list blobs", func() {
					BeforeEach(func() {
						expectedSnapshotCurrent = time.Date(2017, time.January, 02, 01, 01, 01, 01, time.UTC)
						azureClient.ListBlobsCall.Returns.Error = errors.New("failed to list blobs")
					})

					It("returns an error", func() {
						_, err := check.VersionsSince("example.json", time.Now())
						Expect(err).To(MatchError("failed to list blobs"))
					})
				})
			})

			Context("when the file is not found", func() {
				It("returns an error", func() {
					_, err := check.VersionsSince("non-existant.json", time.Now())
					Expect(err).To(MatchError("failed to find blob: non-existant.json"))
				})
			})
		})
	})

	Describe("LatestVersionRegexp", func() {
		Context("given a regex pattern with semver blobs", func() {
			BeforeEach(func() {
				azureClient.ListBlobsCall.Returns.BlobListResponse = storage.BlobListResponse{
					Blobs: []storage.Blob{
						storage.Blob{
							Name: "example-1.0.0.json",
						},
						storage.Blob{
							Name: "example-1.2.3.json",
						},
						storage.Blob{
							Name: "example-1.2.0.json",
						},
						storage.Blob{
							Name: "foo.json",
						},
					},
				}
			})

			It("returns just the latest blob matching the regex pattern", func() {
				latestVersion, err := check.LatestVersionRegexp("example-(.*).json")
				Expect(err).NotTo(HaveOccurred())

				Expect(azureClient.ListBlobsCall.CallCount).To(Equal(1))
				Expect(azureClient.ListBlobsCall.Receives.ListBlobsParameters).To(Equal(storage.ListBlobsParameters{
					Include: &storage.IncludeBlobDataset{
						Snapshots:        true,
						Metadata:         false,
						UncommittedBlobs: false,
						Copy:             true,
					},
				}))

				Expect(latestVersion.Path).To(Equal(stringPtr("example-1.2.3.json")))
				Expect(latestVersion.Version).To(Equal(stringPtr("1.2.3")))
			})
		})

		Context("given a regex pattern with numbered blobs", func() {
			BeforeEach(func() {
				azureClient.ListBlobsCall.Returns.BlobListResponse = storage.BlobListResponse{
					Blobs: []storage.Blob{
						storage.Blob{
							Name: "example-1.json",
						},
						storage.Blob{
							Name: "example-3.json",
						},
						storage.Blob{
							Name: "example-2.json",
						},
						storage.Blob{
							Name: "foo.json",
						},
					},
				}
			})

			It("returns just the latest blob matching the regex pattern", func() {
				latestVersion, err := check.LatestVersionRegexp("example-(.*).json")
				Expect(err).NotTo(HaveOccurred())

				Expect(latestVersion.Path).To(Equal(stringPtr("example-3.json")))
			})
		})

		Context("given a regex pattern with multiple groups", func() {
			BeforeEach(func() {
				azureClient.ListBlobsCall.Returns.BlobListResponse = storage.BlobListResponse{
					Blobs: []storage.Blob{
						storage.Blob{
							Name: "example-a-1.0.0-a.json",
						},
						storage.Blob{
							Name: "example-b-1.2.3-b.json",
						},
						storage.Blob{
							Name: "example-c-1.2.0-c.json",
						},
						storage.Blob{
							Name: "foo.json",
						},
					},
				}
			})

			It("returns a version using the first group as the version", func() {
				latestVersion, err := check.LatestVersionRegexp("example-.-(.*)-(.).json")
				Expect(err).NotTo(HaveOccurred())

				Expect(latestVersion.Path).To(Equal(stringPtr("example-b-1.2.3-b.json")))
			})

			Context("when a group is named version", func() {
				It("returns a version using the named group as the version", func() {
					latestVersion, err := check.LatestVersionRegexp("example-(.)-(?P<version>.*)-..json")
					Expect(err).NotTo(HaveOccurred())

					Expect(latestVersion.Path).To(Equal(stringPtr("example-b-1.2.3-b.json")))
				})
			})
		})

		Context("when no blob is found to match regexp", func() {
			BeforeEach(func() {
				azureClient.ListBlobsCall.Returns.BlobListResponse = storage.BlobListResponse{
					Blobs: []storage.Blob{
						storage.Blob{
							Name: "foo.json",
						},
					},
				}
			})

			It("returns a version using the first group as the version", func() {
				_, err := check.LatestVersionRegexp("example-(.*).json")
				Expect(err).To(MatchError("no matching blob found for regexp: example-(.*).json"))
			})
		})

		Context("when azure client list blobs returns an error", func() {
			BeforeEach(func() {
				azureClient.ListBlobsCall.Returns.Error = errors.New("something bad happened")
			})

			It("returns an error", func() {
				_, err := check.LatestVersionRegexp("example-(.*).json")
				Expect(err).To(MatchError("something bad happened"))
			})
		})

		Context("when an invalid regex pattern is provided", func() {
			It("returns an error", func() {
				_, err := check.LatestVersionRegexp("example-(.json")
				Expect(err).To(MatchError("error parsing regexp: missing closing ): `example-(.json`"))
			})
		})

		Context("when the match is not a valid version number", func() {
			BeforeEach(func() {
				azureClient.ListBlobsCall.Returns.BlobListResponse = storage.BlobListResponse{
					Blobs: []storage.Blob{
						storage.Blob{
							Name: "example-%.json",
						},
					},
				}
			})

			It("returns an error", func() {
				_, err := check.LatestVersionRegexp("example-(.*).json")
				Expect(err).To(MatchError("Expected version '%' to match version format"))
			})
		})
	})
})
