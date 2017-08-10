package api

import "time"

type InRequest struct {
	Source  RequestSource    `json:"source"`
	Version InRequestVersion `json:"version"`
}

type OutRequest struct {
	Params OutParams     `json:"params"`
	Source RequestSource `json:"source"`
}

type RequestSource struct {
	StorageAccountName string `json:"storage_account_name"`
	StorageAccountKey  string `json:"storage_account_key"`
	Container          string `json:"container"`
	VersionedFile      string `json:"versioned_file"`
}

type InRequestVersion struct {
	Snapshot time.Time `json:"snapshot"`
}

type OutParams struct {
	File string `json:"file"`
}
