package models

import (
	"encoding/json"
	"time"
)

type EDDN struct {
	SchemaRef string          `json:"$schemaRef"`
	Header    EDDNHeader      `json:"header"`
	Message   json.RawMessage `json:"message"`
}
type EDDNHeader struct {
	UploaderID       string    `json:"uploaderID"`
	SoftwareName     string    `json:"softwareName"`
	SoftwareVersion  string    `json:"softwareVersion"`
	GatewayTimestamp time.Time `json:"gatewayTimestamp"`
}
