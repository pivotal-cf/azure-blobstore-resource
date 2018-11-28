package api_test

import (
	"bytes"
	"errors"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/pivotal-cf/azure-blobstore-resource/api"
	"github.com/pivotal-cf/azure-blobstore-resource/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("In", func() {
	var (
		azureClient *fakes.AzureClient
		in          api.In

		tempDir string
	)

	BeforeEach(func() {
		azureClient = &fakes.AzureClient{}
		in = api.NewIn(azureClient)

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("CopyBlobToDestination", func() {
		var (
			snapshot time.Time
		)

		BeforeEach(func() {
			azureClient.GetRangeCall.Returns.BlobReader = ioutil.NopCloser(bytes.NewReader([]byte(`{"key": "value"}`)))
			snapshot = time.Date(2017, time.January, 01, 01, 01, 01, 01, time.UTC)

			azureClient.GetBlobSizeInBytesCall.Returns.BlobSize = api.ChunkSize*2 + 50
		})

		It("copies blob from azure blobstore to local destination directory", func() {
			err := in.CopyBlobToDestination(tempDir, "example.json", snapshot)
			Expect(err).NotTo(HaveOccurred())

			Expect(azureClient.GetBlobSizeInBytesCall.CallCount).To(Equal(1))
			Expect(azureClient.GetBlobSizeInBytesCall.Receives.BlobName).To(Equal("example.json"))
			Expect(azureClient.GetBlobSizeInBytesCall.Receives.Snapshot).To(Equal(snapshot))

			Expect(azureClient.GetRangeCall.CallCount).To(Equal(3))
			Expect(azureClient.GetRangeCall.Receives[0].BlobName).To(Equal("example.json"))
			Expect(azureClient.GetRangeCall.Receives[0].Snapshot).To(Equal(snapshot))

			Expect(azureClient.GetRangeCall.Receives[0].StartRangeInBytes).To(Equal(uint64(0)))
			Expect(azureClient.GetRangeCall.Receives[0].EndRangeInBytes).To(Equal(uint64(api.ChunkSize - 1)))

			Expect(azureClient.GetRangeCall.Receives[1].StartRangeInBytes).To(Equal(uint64(api.ChunkSize)))
			Expect(azureClient.GetRangeCall.Receives[1].EndRangeInBytes).To(Equal(uint64(api.ChunkSize*2 - 1)))

			Expect(azureClient.GetRangeCall.Receives[2].StartRangeInBytes).To(Equal(uint64(api.ChunkSize * 2)))
			Expect(azureClient.GetRangeCall.Receives[2].EndRangeInBytes).To(Equal(uint64(api.ChunkSize*2 + 50)))

			data, err := ioutil.ReadFile(filepath.Join(tempDir, "example.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(`{"key": "value"}`))
		})

		Context("when a sub directory is specified within destination", func() {
			It("creates the sub directories", func() {
				err := in.CopyBlobToDestination(tempDir, "./sub/dir/example.json", snapshot)
				Expect(err).NotTo(HaveOccurred())

				Expect(azureClient.GetBlobSizeInBytesCall.CallCount).To(Equal(1))
				Expect(azureClient.GetBlobSizeInBytesCall.Receives.BlobName).To(Equal("./sub/dir/example.json"))
				Expect(azureClient.GetBlobSizeInBytesCall.Receives.Snapshot).To(Equal(snapshot))

				Expect(azureClient.GetRangeCall.CallCount).To(Equal(3))
				Expect(azureClient.GetRangeCall.Receives[0].BlobName).To(Equal("./sub/dir/example.json"))
				Expect(azureClient.GetRangeCall.Receives[0].Snapshot).To(Equal(snapshot))

				Expect(azureClient.GetRangeCall.Receives[0].StartRangeInBytes).To(Equal(uint64(0)))
				Expect(azureClient.GetRangeCall.Receives[0].EndRangeInBytes).To(Equal(uint64(api.ChunkSize - 1)))

				Expect(azureClient.GetRangeCall.Receives[1].StartRangeInBytes).To(Equal(uint64(api.ChunkSize)))
				Expect(azureClient.GetRangeCall.Receives[1].EndRangeInBytes).To(Equal(uint64(api.ChunkSize*2 - 1)))

				Expect(azureClient.GetRangeCall.Receives[2].StartRangeInBytes).To(Equal(uint64(api.ChunkSize * 2)))
				Expect(azureClient.GetRangeCall.Receives[2].EndRangeInBytes).To(Equal(uint64(api.ChunkSize*2 + 50)))

				data, err := ioutil.ReadFile(filepath.Join(tempDir, "sub", "dir", "example.json"))
				Expect(err).NotTo(HaveOccurred())
				Expect(string(data)).To(Equal(`{"key": "value"}`))
			})
		})

		Context("when an error occurs", func() {
			Context("when azure client fails to get a blob", func() {
				It("returns an error", func() {
					azureClient.GetRangeCall.Returns.Error = errors.New("failed to get blob")
					err := in.CopyBlobToDestination(tempDir, "example.json", snapshot)
					Expect(err).To(MatchError("failed to get blob"))
				})
			})

			Context("when it fails to create a file into the destination dir", func() {
				It("returns an error", func() {
					err := in.CopyBlobToDestination("/fake/dest/dir", "example.json", snapshot)
					Expect(err).To(MatchError("mkdir /fake: permission denied"))
				})
			})
		})
	})
})
