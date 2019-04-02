package api

import "time"

type InRequest struct {
	Source  RequestSource    `json:"source"`
	Version InRequestVersion `json:"version"`
	Params  InParams         `json:"params"`
}

type OutRequest struct {
	Params OutParams     `json:"params"`
	Source RequestSource `json:"source"`
}

type RequestSource struct {
	BaseURL            string `json:"base_url"`
	StorageAccountName string `json:"storage_account_name"`
	StorageAccountKey  string `json:"storage_account_key"`
	Container          string `json:"container"`
	VersionedFile      string `json:"versioned_file"`
	Regexp             string `json:"regexp"`
}

type InRequestVersion struct {
	Snapshot time.Time `json:"snapshot,omitempty"`
	Path     string    `json:"path,omitempty"`
	Version  string    `json:"version,omitempty"`
}

type InParams struct {
	Unpack       bool `json:"unpack"`
	SkipDownload bool `json:"skip_download"`
}

type OutParams struct {
	File string `json:"file"`
}
