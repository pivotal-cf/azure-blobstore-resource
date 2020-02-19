package api

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/Azure/azure-sdk-for-go/storage"
	"github.com/cppforlife/go-semi-semantic/version"
)

type Version struct {
	Snapshot *time.Time `json:"snapshot,omitempty"`
	Path     *string    `json:"path,omitempty"`
	Version  *string    `json:"version,omitempty"`

	comparableVersion version.Version
}

type Check struct {
	azureClient azureClient
}

func NewCheck(azureClient azureClient) Check {
	return Check{azureClient: azureClient}
}

func (c Check) VersionsSince(filename string, snapshot time.Time) ([]Version, error) {
	blobs := []storage.Blob{}
	marker := ""

	for {
		blobListResponse, err := c.azureClient.ListBlobs(storage.ListBlobsParameters{
			Prefix: filename,
			Include: &storage.IncludeBlobDataset{
				Snapshots: true,
				Copy:      true,
			},
			Marker: marker,
		})

		if err != nil {
			return []Version{}, err
		}

		for _, blob := range blobListResponse.Blobs {
			blobs = append(blobs, blob)
		}

		marker = blobListResponse.NextMarker
		if marker == "" {
			break
		}
	}

	var newerVersions []Version
	var found bool
	for _, blob := range blobs {
		if blob.Properties.CopyStatus != "" && blob.Properties.CopyStatus != "success" {
			continue // skip blobs which are still being copied
		}

		if blob.Name == filename {
			if blob.Snapshot.After(snapshot) || blob.Snapshot.Equal(snapshot) {
				newerVersions = append(newerVersions, Version{
					Snapshot: timePtr(blob.Snapshot),
				})
			}
			found = true
		}
	}

	if !found {
		return []Version{}, fmt.Errorf("failed to find blob: %s", filename)
	}

	sort.Slice(newerVersions, func(i, j int) bool {
		return newerVersions[i].Snapshot.Before(*newerVersions[j].Snapshot)
	})

	return newerVersions, nil
}

func (c Check) VersionsSinceRegexp(expr, currentVersion string) ([]Version, error) {
	blobs := []storage.Blob{}
	marker := ""

	var hasRan bool
	for {
		blobListResponse, err := c.azureClient.ListBlobs(storage.ListBlobsParameters{
			Include: &storage.IncludeBlobDataset{
				Snapshots: true,
				Copy:      true,
			},
			Marker: marker,
		})

		if err != nil {
			return []Version{}, err
		}

		for _, blob := range blobListResponse.Blobs {
			blobs = append(blobs, blob)
		}

		marker = blobListResponse.NextMarker
		if marker == "" || (hasRan && len(blobListResponse.Blobs) == 0) {
			break
		}

		hasRan = true
	}

	matcher, err := regexp.Compile(expr)
	if err != nil {
		return []Version{}, err
	}

	curVersion, err := version.NewVersionFromString(currentVersion)
	if err != nil {
		// ignored, if currentVersion could not be converted to a version we will
		// assume every version is newer
	}

	var newerVersions []Version
	for _, blob := range blobs {
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
			return []Version{}, err
		}

		if currentVersion == "" || ver.Compare(curVersion) >= 0 {
			newerVersions = append(newerVersions, Version{
				Path:              stringPtr(blob.Name),
				Version:           stringPtr(ver.AsString()),
				comparableVersion: ver,
			})
		}
	}

	if len(newerVersions) == 0 {
		return []Version{}, fmt.Errorf("no matching blob found for regexp: %s", expr)
	}

	sort.Slice(newerVersions, func(i, j int) bool {
		return newerVersions[i].comparableVersion.Compare(newerVersions[j].comparableVersion) < 0
	})

	return newerVersions, nil
}

func findString(items []string, searchFor string) (int, bool) {
	for i, item := range items {
		if item == searchFor {
			return i, true
		}
	}

	return -1, false
}

func stringPtr(str string) *string {
	return &str
}

func timePtr(t time.Time) *time.Time {
	return &t
}
