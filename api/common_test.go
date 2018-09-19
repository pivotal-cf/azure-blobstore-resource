package api_test

import (
	"time"

	. "github.com/pivotal-cf/azure-blobstore-resource/api"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Common", func() {

	Describe("URLAppendTimeStamp", func() {
		var (
			baseURL string
		)

		BeforeEach(func() {
			baseURL = "http://example.com"

		})

		It("keeps trailing zero's in timestamp", func() {
			timestamp := time.Date(2017, 1, 2, 3, 4, 5, 600000*1000, time.UTC)
			url, err := URLAppendTimeStamp(baseURL, timestamp)
			Expect(err).NotTo(HaveOccurred())
			Expect(url).To(Equal("http://example.com?snapshot=2017-01-02T03%3A04%3A05.6000000Z"))
		})

	})

})
