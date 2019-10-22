package api_test

import (
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pivotal-cf/azure-blobstore-resource/azure/azurefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf/azure-blobstore-resource/api"
)

var _ = Describe("Out", func() {
	var (
		azureClient *azurefakes.FakeAzureClient
		out         api.Out

		tempDir string
	)

	BeforeEach(func() {
		azureClient = &azurefakes.FakeAzureClient{}
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
			azureClient.CreateSnapshotReturnsOnCall(0, expectedSnapshot, nil)
		})

		It("uploads blob to azure blobstore from source directory and returns a zero time", func() {
			var expectedStreamData []byte
			azureClient.UploadFromStreamStub = func(blobName string, blockSize int, stream io.Reader) error {
				var err error
				expectedStreamData, err = ioutil.ReadAll(stream)
				Expect(err).NotTo(HaveOccurred())

				return nil
			}

			path, snapshot, err := out.UploadFileToBlobstore(tempDir, "example.json", "example.json", false, 1)
			Expect(err).NotTo(HaveOccurred())

			Expect(azureClient.UploadFromStreamCallCount()).To(Equal(1))

			blobName, blockSize, _ := azureClient.UploadFromStreamArgsForCall(0)
			Expect(blobName).To(Equal("example.json"))
			Expect(blockSize).To(Equal(1))

			Expect(string(expectedStreamData)).To(Equal("some-data"))

			Expect(azureClient.CreateSnapshotCallCount()).To(Equal(0))

			Expect(path).To(Equal("example.json"))
			Expect(snapshot).To(BeNil())
		})

		Context("when a snapshot is desired", func() {
			It("uploads blob to azure blobstore from source directory and returns a snapshot time", func() {
				var expectedStreamData []byte
				azureClient.UploadFromStreamStub = func(blobName string, blockSize int, stream io.Reader) error {
					var err error
					expectedStreamData, err = ioutil.ReadAll(stream)
					Expect(err).NotTo(HaveOccurred())

					return nil
				}

				path, snapshot, err := out.UploadFileToBlobstore(tempDir, "example.json", "some-blob", true, 1)
				Expect(err).NotTo(HaveOccurred())

				Expect(azureClient.UploadFromStreamCallCount()).To(Equal(1))

				blobName, blockSize, _ := azureClient.UploadFromStreamArgsForCall(0)
				Expect(blobName).To(Equal("some-blob"))
				Expect(blockSize).To(Equal(1))

				Expect(string(expectedStreamData)).To(Equal("some-data"))

				Expect(azureClient.CreateSnapshotCallCount()).To(Equal(1))
				Expect(azureClient.CreateSnapshotArgsForCall(0)).To(Equal("some-blob"))

				Expect(path).To(Equal("some-blob"))
				Expect(snapshot).To(Equal(&expectedSnapshot))
			})
		})

		Context("when a glob is provided", func() {
			BeforeEach(func() {
				err := os.Mkdir(filepath.Join(tempDir, "some-sub-dir"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())

				err = ioutil.WriteFile(filepath.Join(tempDir, "some-sub-dir", "example-1.2.json"), []byte("some-data"), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			It("uploads the file that matches the glob to azure blobstore", func() {
				var expectedStreamData []byte
				azureClient.UploadFromStreamStub = func(blobName string, blockSize int, stream io.Reader) error {
					var err error
					expectedStreamData, err = ioutil.ReadAll(stream)
					Expect(err).NotTo(HaveOccurred())

					return nil
				}

				path, snapshot, err := out.UploadFileToBlobstore(tempDir, "some-sub-dir/example-*.json", "some-blob-dir/example-*.json", false, 1)
				Expect(err).NotTo(HaveOccurred())

				Expect(azureClient.UploadFromStreamCallCount()).To(Equal(1))

				blobName, blockSize, _ := azureClient.UploadFromStreamArgsForCall(0)
				Expect(blobName).To(Equal("some-blob-dir/example-1.2.json"))
				Expect(blockSize).To(Equal(1))

				Expect(string(expectedStreamData)).To(Equal("some-data"))

				Expect(azureClient.CreateSnapshotCallCount()).To(Equal(0))

				Expect(path).To(Equal("some-blob-dir/example-1.2.json"))
				Expect(snapshot).To(BeNil())
			})
		})

		Context("when an error occurs", func() {
			Context("when multiple files match", func() {
				BeforeEach(func() {
					err := ioutil.WriteFile(filepath.Join(tempDir, "example-1.2.json"), []byte("some-data"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())

					err = ioutil.WriteFile(filepath.Join(tempDir, "example-1.3.json"), []byte("some-data"), os.ModePerm)
					Expect(err).NotTo(HaveOccurred())
				})

				It("returns an error", func() {
					_, _, err := out.UploadFileToBlobstore(tempDir, "example-*.json", "example-*.json", false, 1)
					Expect(err).To(MatchError("multiple files match glob: example-*.json"))
				})
			})

			Context("when it fails to open file", func() {
				It("returns an error", func() {
					_, _, err := out.UploadFileToBlobstore("/fake/source/dir", "example.json", "example.json", false, 1)
					Expect(err).To(MatchError("open /fake/source/dir/example.json: no such file or directory"))
				})
			})

			Context("when azure client fails to upload from stream", func() {
				It("returns an error", func() {
					azureClient.UploadFromStreamReturnsOnCall(0, errors.New("failed to upload blob"))
					_, _, err := out.UploadFileToBlobstore(tempDir, "example.json", "example.json", false, 1)
					Expect(err).To(MatchError("failed to upload blob"))
				})
			})

			Context("when azure client fails to create snapshot", func() {
				It("returns an error", func() {
					azureClient.CreateSnapshotReturnsOnCall(0, time.Now(), errors.New("failed to create snapshot"))
					_, _, err := out.UploadFileToBlobstore(tempDir, "example.json", "example.json", true, 1)
					Expect(err).To(MatchError("failed to create snapshot"))
				})
			})
		})
	})
})
