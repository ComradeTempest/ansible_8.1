// Exchange data structures are described here
package main

// Get start position request structure
type SRequestStartPos struct {
	Host     string `json:"host"`
	Instance string `json:"instance"`
}

// Process CDR request structure
type SProcessCdr struct {
	Host         string `json:"host"`
	Instance     string `json:"instance"`
	Filename     string `json:"filename"`
	Position     int64  `json:"position"`
	RecordLength int    `json:"recordLength"`
	Data         string `json:"data"`
}

// Get start position and process CDR response JSON structure
type SResponseJson struct {
	Filename string `json:"startFile"`
	Position int64  `json:"startPosition"`
}
