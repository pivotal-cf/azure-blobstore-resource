package api

import (
	"fmt"
	"regexp"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/cppforlife/go-semi-semantic/version"
)

type Version struct {
	Snapshot *time.Time `json:"snapshot,omitempty"`
	Path     *string    `json:"path,omitempty"`
	Version  *string    `json:"version,omitempty"`
}

type Check struct {
	azureClient azureClient
}

func NewCheck(azureClient azureClient) Check {
	return Check{azureClient: azureClient}
}

func (c Check) LatestVersion(filename string) (Version, error) {
	blobListResponse, err := c.azureClient.ListBlobs(storage.ListBlobsParameters{
		Prefix: filename,
		Include: &storage.IncludeBlobDataset{
			Snapshots: true,
			Copy:      true,
		},
	})
	if err != nil {
		return Version{}, err
	}

	var latestSnapshot time.Time
	var found bool
	for _, blob := range blobListResponse.Blobs {
		if blob.Properties.CopyStatus != "" && blob.Properties.CopyStatus != "success" {
			continue // skip blobs which are still being copied
		}

		if blob.Name == filename {
			if blob.Snapshot.After(latestSnapshot) {
				latestSnapshot = blob.Snapshot
			}
			found = true
		}
	}

	if !found {
		return Version{}, fmt.Errorf("failed to find blob: %s", filename)
	}

	return Version{
		Snapshot: &latestSnapshot,
	}, nil
}

func (c Check) LatestVersionRegexp(expr string) (Version, error) {
	blobListResponse, err := c.azureClient.ListBlobs(storage.ListBlobsParameters{
		Include: &storage.IncludeBlobDataset{
			Snapshots: true,
			Copy:      true,
		},
	})
	if err != nil {
		return Version{}, err
	}

	matcher, err := regexp.Compile(expr)
	if err != nil {
		return Version{}, err
	}

	var latestVersion version.Version
	var latestBlobName string
	for _, blob := range blobListResponse.Blobs {
		if blob.Properties.CopyStatus != "" && blob.Properties.CopyStatus != "success" {
			continue // skip blobs which are still being copied
		}

		var match string

		matches := matcher.FindStringSubmatch(blob.Name)
		if len(matches) < 2 {
			continue // no match
		}

		index, found := findString(matcher.SubexpNames(), "version")
		if found {
			match = matches[index]
		} else {
			match = matches[1]
		}

		ver, err := version.NewVersionFromString(match)
		if err != nil {
			return Version{}, err
		}

		if ver.Compare(latestVersion) >= 0 {
			latestVersion = ver
			latestBlobName = blob.Name
		}
	}

	if latestBlobName == "" {
		return Version{}, fmt.Errorf("no matching blob found for regexp: %s", expr)
	}

	version := latestVersion.AsString()
	return Version{
		Path:    &latestBlobName,
		Version: &version,
	}, nil
}

func findString(items []string, searchFor string) (int, bool) {
	for i, item := range items {
		if item == searchFor {
			return i, true
		}
	}

	return -1, false
}
