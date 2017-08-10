package api

import "time"

type InRequest struct {
	Source  InRequestSource  `json:"source"`
	Version InRequestVersion `json:"version"`
}

type InRequestSource struct {
	StorageAccountName string `json:"storage_account_name"`
	StorageAccountKey  string `json:"storage_account_key"`
	Container          string `json:"container"`
	VersionedFile      string `json:"versioned_file"`
}

type InRequestVersion struct {
	Snapshot time.Time `json:"snapshot"`
}
