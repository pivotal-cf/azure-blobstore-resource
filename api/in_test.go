package api_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"path"
	"path/filepath"
	"time"

	"github.com/pivotal-cf/azure-blobstore-resource/azure/azurefakes"

	"github.com/pivotal-cf/azure-blobstore-resource/api"

	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("In", func() {
	var (
		azureClient *azurefakes.FakeAzureClient
		in          api.In

		tempDir string
	)

	BeforeEach(func() {
		azureClient = &azurefakes.FakeAzureClient{}
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
			snapshot = time.Date(2017, time.January, 01, 01, 01, 01, 01, time.UTC)
		})

		It("copies blob from azure blobstore to local destination directory", func() {
			err := in.CopyBlobToDestination(tempDir, "example.json", &snapshot, 1, time.Second)
			Expect(err).NotTo(HaveOccurred())

			Expect(azureClient.DownloadBlobToFileCallCount()).To(Equal(1))

			blobName, file, passedSnapshot, blockSize, retryTryTimeout := azureClient.DownloadBlobToFileArgsForCall(0)
			Expect(blobName).To(Equal("example.json"))
			Expect(path.Base(file.Name())).To(Equal("example.json"))
			Expect(passedSnapshot).To(Equal(&snapshot))
			Expect(blockSize).To(Equal(int64(1)))
			Expect(retryTryTimeout).To(Equal(time.Second))
		})

		Context("when a sub directory is specified within destination", func() {
			It("does not create the sub directories (matches s3 resource implementation)", func() {
				err := in.CopyBlobToDestination(tempDir, "./sub/dir/example.json", &snapshot, 1, time.Second)
				Expect(err).NotTo(HaveOccurred())

				Expect(azureClient.DownloadBlobToFileCallCount()).To(Equal(1))

				blobName, file, passedSnapshot, blockSize, retryTryTimeout := azureClient.DownloadBlobToFileArgsForCall(0)
				Expect(blobName).To(Equal("./sub/dir/example.json"))
				Expect(path.Base(file.Name())).To(Equal("example.json"))
				Expect(passedSnapshot).To(Equal(&snapshot))
				Expect(blockSize).To(Equal(int64(1)))
				Expect(retryTryTimeout).To(Equal(time.Second))
			})
		})

		Context("when an error occurs", func() {
			Context("when azure client fails to get a blob", func() {
				It("returns an error", func() {
					azureClient.DownloadBlobToFileReturns(errors.New("failed to get blob"))
					err := in.CopyBlobToDestination(tempDir, "example.json", &snapshot, 1, time.Second)
					Expect(err).To(MatchError("failed to get blob"))
				})
			})
		})
	})

	Describe("UnpackBlob", func() {
		DescribeTable("unpacks the blob successfully", func(fixtureFilename, innerFilename, innerFileContents string) {
			err := copyFile(filepath.Join("fixtures", fixtureFilename), filepath.Join(tempDir, fixtureFilename))
			Expect(err).NotTo(HaveOccurred())

			err = in.UnpackBlob(filepath.Join(tempDir, fixtureFilename))
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(filepath.Join(tempDir, innerFilename))
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(filepath.Join(tempDir, fixtureFilename))
			Expect(err.Error()).To(ContainSubstring("no such file or directory"))

			body, err := ioutil.ReadFile(filepath.Join(tempDir, innerFilename))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(body)).To(ContainSubstring(innerFileContents))
		},
			Entry("when the blob is a tarball", "example.tar", filepath.Join("example", "foo.txt"), "gopher"),
			Entry("when the blob is a tgz", "example.tgz", filepath.Join("example", "foo.txt"), "gopher"),
			Entry("when the blob is a tar.gz", "example.tar.gz", filepath.Join("example", "foo.txt"), "gopher"),
			Entry("when the blob is a zip", "example.zip", filepath.Join("example", "foo.txt"), "gopher"),
		)

		Context("when an invalid archive is provided", func() {
			It("returns an error", func() {
				err := copyFile(filepath.Join("fixtures", "example.txt"), filepath.Join(tempDir, "example.txt"))
				Expect(err).NotTo(HaveOccurred())

				err = in.UnpackBlob(filepath.Join(tempDir, "example.txt"))
				Expect(err).To(MatchError(fmt.Sprintf("invalid archive: %s", filepath.Join(tempDir, "example.txt"))))
			})
		})

		It("returns an error when un-tar fails", func() {
			err := in.UnpackBlob("does-not-exist.tgz")
			Expect(err).To(MatchError("open does-not-exist.tgz: no such file or directory"))
		})
	})
})

func copyFile(sourceFilename, destinationFilename string) error {
	src, err := os.Open(sourceFilename)
	if err != nil {
		return err
	}

	dst, err := os.OpenFile(destinationFilename, os.O_RDWR|os.O_CREATE, os.ModePerm)
	if err != nil {
		return err
	}

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	return nil
}
