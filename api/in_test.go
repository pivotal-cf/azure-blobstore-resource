package api_test

import (
	"errors"
	"io/ioutil"
	"path/filepath"
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

		BeforeEach(func() {
			azureClient.GetCall.Returns.BlobData = []byte(`{"key": "value"}`)
		})

		It("copies blob from azure blobstore to local destination directory", func() {
			err := in.CopyBlobToDestination(tempDir, "example.json")
			Expect(err).NotTo(HaveOccurred())

			Expect(azureClient.GetCall.CallCount).To(Equal(1))
			Expect(azureClient.GetCall.Receives.BlobName).To(Equal("example.json"))

			data, err := ioutil.ReadFile(filepath.Join(tempDir, "example.json"))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(data)).To(Equal(`{"key": "value"}`))
		})

		Context("when an error occurs", func() {
			Context("when azure client fails to get a blob", func() {
				It("returns an error", func() {
					azureClient.GetCall.Returns.Error = errors.New("failed to get blob")
					err := in.CopyBlobToDestination(tempDir, "example.json")
					Expect(err).To(MatchError("failed to get target blob: failed to get blob"))
				})
			})

			Context("when it fails to write a file into the destination dir", func() {
				It("returns an error", func() {
					err := in.CopyBlobToDestination("/fake/dest/dir", "example.json")
					Expect(err).To(MatchError("failed to write target blob to /fake/dest/dir/example.json: open /fake/dest/dir/example.json: no such file or directory"))
				})
			})
		})
	})
})
