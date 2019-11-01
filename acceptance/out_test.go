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

var _ = Describe("Out", func() {
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

	Context("when given a source directory with updated blob", func() {
		BeforeEach(func() {
			err := os.Mkdir(filepath.Join(tempDir, "some-resource"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tempDir, "some-resource", "example.json"), []byte("updated"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("uploads the blob to azure blobstore from the source directory", func() {
			out := exec.Command(pathToOut, tempDir)
			out.Stderr = os.Stderr

			stdin, err := out.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			_, err = io.WriteString(stdin, fmt.Sprintf(`{
					"params": {
						"file": "some-resource/example.json"
					},
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

			outputJSON, err := out.Output()
			Expect(err).NotTo(HaveOccurred())

			var output struct {
				Version struct {
					Snapshot time.Time `json:"snapshot"`
				} `json:"version"`
			}
			err = json.Unmarshal(outputJSON, &output)
			Expect(err).NotTo(HaveOccurred())

			data := downloadBlobWithSnapshot(container, "example.json", output.Version.Snapshot)
			Expect(string(data)).To(Equal("updated"))
		})
	})

	Context("when given a source directory with a very large blob", func() {
		BeforeEach(func() {
			err := os.Mkdir(filepath.Join(tempDir, "some-resource"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = makeLargeFile(filepath.Join(tempDir, "some-resource", "big_file"), 200*1000*1000)
			Expect(err).NotTo(HaveOccurred())
		})

		It("uploads the very large blob to azure blobstore from the source directory", func() {
			out := exec.Command(pathToOut, tempDir)
			out.Stderr = os.Stderr

			stdin, err := out.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			_, err = io.WriteString(stdin, fmt.Sprintf(`{
					"params": {
						"file": "some-resource/big_file",
						"retry": {
							"try_timeout": "1m"
						}
					},
					"source": {
						"storage_account_name": %q,
						"storage_account_key": %q,
						"container": %q,
						"versioned_file": "big_file"
					}
				}`,
				config.StorageAccountName,
				config.StorageAccountKey,
				container,
			))
			Expect(err).NotTo(HaveOccurred())

			outputJSON, err := out.Output()
			Expect(err).NotTo(HaveOccurred())

			var output struct {
				Version struct {
					Snapshot time.Time `json:"snapshot"`
				} `json:"version"`
			}
			err = json.Unmarshal(outputJSON, &output)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	Context("when given a regex pattern is provided", func() {
		BeforeEach(func() {
			err := os.Mkdir(filepath.Join(tempDir, "some-resource"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = os.Mkdir(filepath.Join(tempDir, "some-resource", "some-sub-dir"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())

			err = ioutil.WriteFile(filepath.Join(tempDir, "some-resource", "some-sub-dir", "example-1.2.txt"), []byte("updated"), os.ModePerm)
			Expect(err).NotTo(HaveOccurred())
		})

		It("uploads the blob to azure blobstore that matches the file param", func() {
			out := exec.Command(pathToOut, tempDir)
			out.Stderr = os.Stderr

			stdin, err := out.StdinPipe()
			Expect(err).NotTo(HaveOccurred())

			_, err = io.WriteString(stdin, fmt.Sprintf(`{
					"params": {
						"file": "some-resource/some-sub-dir/example-*.txt"
					},
					"source": {
						"storage_account_name": %q,
						"storage_account_key": %q,
						"container": %q,
						"regexp": "some-blob-sub-dir/example-(.*).txt"
					}
				}`,
				config.StorageAccountName,
				config.StorageAccountKey,
				container,
			))
			Expect(err).NotTo(HaveOccurred())

			outputJSON, err := out.Output()
			Expect(err).NotTo(HaveOccurred())

			var output struct {
				Version struct {
					Path string `json:"path"`
				} `json:"version"`
			}
			err = json.Unmarshal(outputJSON, &output)
			Expect(err).NotTo(HaveOccurred())

			Expect(output.Version.Path).To(Equal("some-blob-sub-dir/example-1.2.txt"))
		})
	})
})

func makeLargeFile(filename string, size int64) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}

	err = file.Truncate(size)
	if err != nil {
		return err
	}

	return nil
}
