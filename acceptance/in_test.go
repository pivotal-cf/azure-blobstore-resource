package acceptance_test

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("In", func() {
	var (
		container string
		tempDir   string
	)

	BeforeEach(func() {
		rand.Seed(time.Now().UTC().UnixNano())
		container = fmt.Sprintf("azureblobstore%d", rand.Int())
		createContainer(container)

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

	})

	AfterEach(func() {
		deleteContainer(container)
	})

	Context("when given a specific snapshot version and destination directory", func() {
		var (
			snapshotTimestamp *time.Time
		)

		BeforeEach(func() {
			snapshotTimestamp = createBlobWithSnapshot(container, "example.json")
		})

		It("downloads the specific blob version and copies it to destination directory", func() {
			in := exec.Command(pathToIn, tempDir)
			in.Stderr = os.Stderr

			stdin, err := in.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			_, err = io.WriteString(stdin, fmt.Sprintf(`{
					"source": {
						"storage_account_name": %q,
						"storage_account_key": %q,
						"container": %q,
						"versioned_file": "example.json"
					},
					"version": { "snapshot": %q }
				}`,
				config.StorageAccountName,
				config.StorageAccountKey,
				container,
				snapshotTimestamp.Format(time.RFC3339Nano),
			))
			Expect(err).NotTo(HaveOccurred())

			outputJSON, err := in.Output()
			Expect(err).NotTo(HaveOccurred())

			var output struct {
				Version struct {
					Snapshot time.Time `json:"snapshot"`
				} `json:"version"`
				Metadata []struct {
					Name  string `json:"name"`
					Value string `json:"value"`
				} `json:"metadata"`
			}
			err = json.Unmarshal(outputJSON, &output)
			Expect(err).NotTo(HaveOccurred())

			Expect(output.Version.Snapshot).To(Equal(*snapshotTimestamp))
			Expect(output.Metadata[0].Name).To(Equal("filename"))
			Expect(output.Metadata[0].Value).To(Equal("example.json"))
			Expect(output.Metadata[1].Name).To(Equal("url"))
			Expect(output.Metadata[1].Value).To(Equal(fmt.Sprintf("https://%s.blob.core.windows.net/%s/example.json", config.StorageAccountName, container)))

			_, err = os.Stat(filepath.Join(tempDir, "example.json"))
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
