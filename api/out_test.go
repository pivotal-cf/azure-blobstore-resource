package api_test

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/christianang/azure-blobstore-resource/api"
	"github.com/christianang/azure-blobstore-resource/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Out", func() {
	var (
		azureClient *fakes.AzureClient
		out         api.Out

		tempDir string
	)

	BeforeEach(func() {
		azureClient = &fakes.AzureClient{}
		out = api.NewOut(azureClient)

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		err = ioutil.WriteFile(filepath.Join(tempDir, "example.json"), []byte("some-data"), os.ModePerm)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("UploadBlobToBlobstore", func() {
		var (
			expectedSnapshot time.Time
		)

		BeforeEach(func() {
			expectedSnapshot = time.Date(2017, time.January, 01, 01, 01, 01, 01, time.UTC)
			azureClient.CreateSnapshotCall.Returns.Snapshot = expectedSnapshot
		})

		It("uploads blob to azure blobstore from source directory and returns a snapshot time", func() {
			var expectedStreamData []byte
			azureClient.UploadFromStreamCall.Stub = func(blobName string, stream io.Reader) error {
				var err error
				expectedStreamData, err = ioutil.ReadAll(stream)
				Expect(err).NotTo(HaveOccurred())

				return nil
			}

			snapshot, err := out.UploadFileToBlobstore(tempDir, "example.json", "example.json")
			Expect(err).NotTo(HaveOccurred())

			Expect(azureClient.UploadFromStreamCall.CallCount).To(Equal(1))
			Expect(azureClient.UploadFromStreamCall.Receives.BlobName).To(Equal("example.json"))
			Expect(string(expectedStreamData)).To(Equal("some-data"))

			Expect(azureClient.CreateSnapshotCall.CallCount).To(Equal(1))
			Expect(azureClient.CreateSnapshotCall.Receives.BlobName).To(Equal("example.json"))

			Expect(snapshot).To(Equal(expectedSnapshot))
		})

		Context("when an error occurs", func() {
			Context("when it fails to open file", func() {
				It("returns an error", func() {
					_, err := out.UploadFileToBlobstore("/fake/source/dir", "example.json", "example.json")
					Expect(err).To(MatchError("open /fake/source/dir/example.json: no such file or directory"))
				})
			})

			Context("when azure client fails to upload from stream", func() {
				It("returns an error", func() {
					azureClient.UploadFromStreamCall.Returns.Error = errors.New("failed to upload blob")
					_, err := out.UploadFileToBlobstore(tempDir, "example.json", "example.json")
					Expect(err).To(MatchError("failed to upload blob"))
				})
			})

			Context("when azure client fails to create snapshot", func() {
				It("returns an error", func() {
					azureClient.CreateSnapshotCall.Returns.Error = errors.New("failed to create snapshot")
					_, err := out.UploadFileToBlobstore(tempDir, "example.json", "example.json")
					Expect(err).To(MatchError("failed to create snapshot"))
				})
			})
		})
	})
})
