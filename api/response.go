package api

import "time"

type Response struct {
	Version  ResponseVersion    `json:"version"`
	Metadata []ResponseMetadata `json:"metadata"`
}

type ResponseVersion struct {
	Snapshot time.Time `json:"snapshot"`
}

type ResponseMetadata struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
