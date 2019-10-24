package acceptance_test

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	SnapshotTimeFormat = "2006-01-02T15:04:05.0000000Z"
)

func TestAcceptance(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "acceptance")
}

type Config struct {
	StorageAccountName string
	StorageAccountKey  string
}

var (
	pathToCheck string
	pathToIn    string
	pathToOut   string
	config      Config
)

var _ = BeforeSuite(func() {
	var err error

	pathToCheck, err = gexec.Build("github.com/pivotal-cf/azure-blobstore-resource/cmd/check")
	Expect(err).NotTo(HaveOccurred())

	pathToIn, err = gexec.Build("github.com/pivotal-cf/azure-blobstore-resource/cmd/in")
	Expect(err).NotTo(HaveOccurred())

	pathToOut, err = gexec.Build("github.com/pivotal-cf/azure-blobstore-resource/cmd/out")
	Expect(err).NotTo(HaveOccurred())

	config = loadConfig()
})

var _ = AfterSuite(func() {
	gexec.CleanupBuildArtifacts()
})

func loadConfig() Config {
	config := Config{
		StorageAccountName: os.Getenv("TEST_STORAGE_ACCOUNT_NAME"),
		StorageAccountKey:  os.Getenv("TEST_STORAGE_ACCOUNT_KEY"),
	}

	if config.StorageAccountName == "" {
		log.Fatal("expected TEST_STORAGE_ACCOUNT_NAME to be set")
	}

	if config.StorageAccountKey == "" {
		log.Fatal("expected TEST_STORAGE_ACCOUNT_KEY to be set")
	}

	return config
}

func createContainer(container string) {
	client, err := storage.NewBasicClient(os.Getenv("TEST_STORAGE_ACCOUNT_NAME"), os.Getenv("TEST_STORAGE_ACCOUNT_KEY"))
	Expect(err).NotTo(HaveOccurred())

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(container)
	err = cnt.Create(&storage.CreateContainerOptions{
		Access: storage.ContainerAccessTypePrivate,
	})
	Expect(err).NotTo(HaveOccurred())
}

func deleteContainer(container string) {
	client, err := storage.NewBasicClient(os.Getenv("TEST_STORAGE_ACCOUNT_NAME"), os.Getenv("TEST_STORAGE_ACCOUNT_KEY"))
	Expect(err).NotTo(HaveOccurred())

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(container)
	err = cnt.Delete(&storage.DeleteContainerOptions{})
	Expect(err).NotTo(HaveOccurred())
}

func createBlobWithSnapshot(container, blobName string) *time.Time {
	client, err := storage.NewBasicClient(os.Getenv("TEST_STORAGE_ACCOUNT_NAME"), os.Getenv("TEST_STORAGE_ACCOUNT_KEY"))
	Expect(err).NotTo(HaveOccurred())

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(container)
	blob := cnt.GetBlobReference(blobName)
	err = blob.CreateBlockBlob(&storage.PutBlobOptions{})
	Expect(err).NotTo(HaveOccurred())

	timestamp, err := blob.CreateSnapshot(&storage.SnapshotOptions{})
	Expect(err).NotTo(HaveOccurred())

	return timestamp
}

func createBlobWithSnapshotAndData(container, blobName, data string) *time.Time {
	client, err := storage.NewBasicClient(os.Getenv("TEST_STORAGE_ACCOUNT_NAME"), os.Getenv("TEST_STORAGE_ACCOUNT_KEY"))
	Expect(err).NotTo(HaveOccurred())

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(container)
	blob := cnt.GetBlobReference(blobName)
	err = blob.CreateBlockBlob(&storage.PutBlobOptions{})
	Expect(err).NotTo(HaveOccurred())

	blockID := base64.StdEncoding.EncodeToString([]byte("BlockID{0000001}"))
	err = blob.PutBlock(blockID, []byte(data), &storage.PutBlockOptions{})

	err = blob.PutBlockList([]storage.Block{
		{
			blockID,
			storage.BlockStatusUncommitted,
		},
	}, &storage.PutBlockListOptions{})
	Expect(err).NotTo(HaveOccurred())

	timestamp, err := blob.CreateSnapshot(&storage.SnapshotOptions{})
	Expect(err).NotTo(HaveOccurred())

	return timestamp
}

func createBlob(container, blobName string) {
	client, err := storage.NewBasicClient(os.Getenv("TEST_STORAGE_ACCOUNT_NAME"), os.Getenv("TEST_STORAGE_ACCOUNT_KEY"))
	Expect(err).NotTo(HaveOccurred())

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(container)
	blob := cnt.GetBlobReference(blobName)
	err = blob.CreateBlockBlob(&storage.PutBlobOptions{})
	Expect(err).NotTo(HaveOccurred())
}

func copyBlob(container, blobName, sourceUrl string) {
	client, err := storage.NewBasicClient(os.Getenv("TEST_STORAGE_ACCOUNT_NAME"), os.Getenv("TEST_STORAGE_ACCOUNT_KEY"))
	Expect(err).NotTo(HaveOccurred())

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(container)
	blob := cnt.GetBlobReference(blobName)
	go func() {
		blob.Copy(sourceUrl, &storage.CopyOptions{})
	}()
}

func uploadBlobWithSnapshot(container, blobName, filename string) *time.Time {
	client, err := storage.NewBasicClient(os.Getenv("TEST_STORAGE_ACCOUNT_NAME"), os.Getenv("TEST_STORAGE_ACCOUNT_KEY"))
	Expect(err).NotTo(HaveOccurred())

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(container)
	blob := cnt.GetBlobReference(blobName)

	err = blob.CreateBlockBlob(&storage.PutBlobOptions{})
	Expect(err).NotTo(HaveOccurred())

	file, err := os.Open(filename)
	Expect(err).NotTo(HaveOccurred())
	defer file.Close()

	var chunkSize = 4000000 // 4Mb
	buffer := make([]byte, chunkSize)
	blocks := []storage.Block{}
	i := 0
	for {
		bytesRead, err := file.Read(buffer)
		if err != nil {
			if err == io.EOF {
				break
			}

			Expect(err).NotTo(HaveOccurred())
		}

		chunk := buffer[:bytesRead]
		blockID := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("BlockID{%07d}", i)))
		err = blob.PutBlock(blockID, chunk, &storage.PutBlockOptions{})
		Expect(err).NotTo(HaveOccurred())

		blocks = append(blocks, storage.Block{
			blockID,
			storage.BlockStatusUncommitted,
		})

		i++
	}

	err = blob.PutBlockList(blocks, &storage.PutBlockListOptions{})
	Expect(err).NotTo(HaveOccurred())

	timestamp, err := blob.CreateSnapshot(&storage.SnapshotOptions{})
	Expect(err).NotTo(HaveOccurred())

	return timestamp
}

func downloadBlobWithSnapshot(container, blobName string, snapshot time.Time) []byte {
	client, err := storage.NewBasicClient(os.Getenv("TEST_STORAGE_ACCOUNT_NAME"), os.Getenv("TEST_STORAGE_ACCOUNT_KEY"))
	Expect(err).NotTo(HaveOccurred())

	blobClient := client.GetBlobService()
	cnt := blobClient.GetContainerReference(container)
	blob := cnt.GetBlobReference(blobName)

	blobReader, err := blob.Get(&storage.GetBlobOptions{
		Snapshot: &snapshot,
	})
	Expect(err).NotTo(HaveOccurred())
	defer blobReader.Close()

	data, err := ioutil.ReadAll(blobReader)
	Expect(err).NotTo(HaveOccurred())

	return data
}
