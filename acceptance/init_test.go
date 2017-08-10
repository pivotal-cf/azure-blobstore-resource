package acceptance_test

import (
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

	pathToCheck, err = gexec.Build("github.com/christianang/azure-blobstore-resource/cmd/check")
	Expect(err).NotTo(HaveOccurred())

	pathToIn, err = gexec.Build("github.com/christianang/azure-blobstore-resource/cmd/in")
	Expect(err).NotTo(HaveOccurred())

	pathToOut, err = gexec.Build("github.com/christianang/azure-blobstore-resource/cmd/out")
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
