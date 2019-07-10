package acceptance_test

import (
	"bytes"
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

		It("returns just the latest blob snapshot version", func() {
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
					},
					"version": { "snapshot": "2017-08-08T23:27:16.2942812Z" }
				}`,
				config.StorageAccountName,
				config.StorageAccountKey,
				container,
			))
			Expect(err).NotTo(HaveOccurred())

			output, err := check.Output()
			Expect(err).NotTo(HaveOccurred())

			var versions []struct {
				Path     *string    `json:"path"`
				Snapshot *time.Time `json:"snapshot"`
			}
			err = json.Unmarshal(output, &versions)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(versions)).To(Equal(1))
			Expect(versions[0].Path).To(BeNil())
			Expect(versions[0].Snapshot).To(Equal(snapshotTimestamp))
		})
	})

	Context("when blob doesn't have a snapshot", func() {
		BeforeEach(func() {
			createBlob(container, "example.json")
		})

		It("returns a zero timestamp version", func() {
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
				Path     *string    `json:"path"`
				Snapshot *time.Time `json:"snapshot"`
			}
			err = json.Unmarshal(output, &versions)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(versions)).To(Equal(1))
			Expect(versions[0].Snapshot).To(Equal(&time.Time{}))
			Expect(versions[0].Path).To(BeNil())
		})
	})

	Context("when there is no blob", func() {
		It("returns an error", func() {
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
					},
					"version": { "snapshot": "2017-08-08T23:27:16.2942812Z" }
				}`,
				config.StorageAccountName,
				config.StorageAccountKey,
				container,
			))
			Expect(err).NotTo(HaveOccurred())

			var stderr bytes.Buffer
			check.Stderr = &stderr

			err = check.Run()
			Expect(err).To(HaveOccurred())

			Expect(stderr.String()).To(ContainSubstring("failed to find blob: example.json"))
		})
	})

	Context("when a regex pattern is provided", func() {
		BeforeEach(func() {
			createBlob(container, "example-1.2.3.json")
		})

		It("returns just the latest version that matches the regexp", func() {
			check := exec.Command(pathToCheck)
			check.Stderr = os.Stderr

			stdin, err := check.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			_, err = io.WriteString(stdin, fmt.Sprintf(`{
					"source": {
						"storage_account_name": %q,
						"storage_account_key": %q,
						"container": %q,
						"regexp": "example-(.*).json"
					},
					"version": { "path": "1.0.0" }
				}`,
				config.StorageAccountName,
				config.StorageAccountKey,
				container,
			))
			Expect(err).NotTo(HaveOccurred())

			output, err := check.Output()
			Expect(err).NotTo(HaveOccurred())

			var versions []struct {
				Path     *string    `json:"path"`
				Snapshot *time.Time `json:"snapshot"`
			}
			err = json.Unmarshal(output, &versions)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(versions)).To(Equal(1))
			Expect(versions[0].Path).To(Equal(stringPtr("example-1.2.3.json")))
			Expect(versions[0].Snapshot).To(BeNil())
		})
	})

	Context("when a blob is being copied", func() {
		BeforeEach(func() {
			createBlob(container, "example-1.2.3.json")
		})

		It("returns just the latest version that matches the regexp which has been copied", func() {
			copyBlob(container, "example-2.3.4.json", "http://example.com")

			Eventually(func() *string {
				check := exec.Command(pathToCheck)
				check.Stderr = os.Stderr

				stdin, err := check.StdinPipe()
				Expect(err).NotTo(HaveOccurred())

				_, err = io.WriteString(stdin, fmt.Sprintf(`{
					"source": {
						"storage_account_name": %q,
						"storage_account_key": %q,
						"container": %q,
						"regexp": "example-(.*).json"
					},
					"version": { "path": "1.0.0" }
				}`,
					config.StorageAccountName,
					config.StorageAccountKey,
					container,
				))
				Expect(err).NotTo(HaveOccurred())

				output, err := check.Output()
				Expect(err).NotTo(HaveOccurred())

				var versions []struct {
					Path     *string    `json:"path"`
					Snapshot *time.Time `json:"snapshot"`
				}
				err = json.Unmarshal(output, &versions)
				Expect(err).NotTo(HaveOccurred())
				return versions[0].Path
			}, 10*time.Second, time.Second).Should(Equal(stringPtr("example-2.3.4.json")))
		})

		It("returns just the latest version that matches the regexp which has been copied", func() {
			copyBlob(container, "example-2.3.4.json", "http://does.not.exist")

			Consistently(func() *string {
				check := exec.Command(pathToCheck)
				check.Stderr = os.Stderr

				stdin, err := check.StdinPipe()
				Expect(err).NotTo(HaveOccurred())

				_, err = io.WriteString(stdin, fmt.Sprintf(`{
					"source": {
						"storage_account_name": %q,
						"storage_account_key": %q,
						"container": %q,
						"regexp": "example-(.*).json"
					},
					"version": { "path": "1.0.0" }
				}`,
					config.StorageAccountName,
					config.StorageAccountKey,
					container,
				))
				Expect(err).NotTo(HaveOccurred())

				output, err := check.Output()
				Expect(err).NotTo(HaveOccurred())

				var versions []struct {
					Path     *string    `json:"path"`
					Snapshot *time.Time `json:"snapshot"`
				}
				err = json.Unmarshal(output, &versions)
				Expect(err).NotTo(HaveOccurred())
				return versions[0].Path
			}).Should(Equal(stringPtr("example-1.2.3.json")))
		})
	})
})

func stringPtr(value string) *string {
	return &value
}
