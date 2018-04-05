package acceptance_test

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Check", func() {
	var (
		container string
	)

	BeforeEach(func() {
		rand.Seed(time.Now().UTC().UnixNano())
		container = fmt.Sprintf("azureblobstore%d", rand.Int())
		createContainer(container)
	})

	AfterEach(func() {
		deleteContainer(container)
	})

	Context("when given any version", func() {
		var (
			snapshotTimestamp *time.Time
		)

		BeforeEach(func() {
			snapshotTimestamp = createBlobWithSnapshot(container, "example.json")
		})

		It("returns just the latest blob", func() {
			check := exec.Command(pathToCheck)
			check.Stderr = os.Stderr

			stdin, err := check.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			_, err = io.WriteString(stdin, fmt.Sprintf(`{
					"source": {
						"storage_account_name": %q,
						"storage_account_key": %q,
						"container": %q,
						"versioned_file": "example.json"
					}
				}`,
				config.StorageAccountName,
				config.StorageAccountKey,
				container,
			))
			Expect(err).NotTo(HaveOccurred())

			output, err := check.Output()
			Expect(err).NotTo(HaveOccurred())

			var versions []struct {
				Snapshot time.Time `json:"snapshot"`
			}
			err = json.Unmarshal(output, &versions)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(versions)).To(Equal(1))
			Expect(versions[0].Snapshot).To(Equal(*snapshotTimestamp))
		})
	})
})
