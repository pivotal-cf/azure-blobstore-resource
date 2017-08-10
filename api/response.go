package api

import "time"

type Response struct {
	Version ResponseVersion `json:"version"`
}

type ResponseVersion struct {
	Snapshot time.Time `json:"snapshot"`
}
