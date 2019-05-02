package registry

import "encoding/json"

type Manifest struct {
	SchemaVersion int     `json:"schemaVersion"`
	MediaType     string  `json:"mediaType"`
	Layers        []Layer `json:"layers"`
}

type ManifestConfig struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

type Layer struct {
	MediaType string `json:"mediaType"`
	Size      int    `json:"size"`
	Digest    string `json:"digest"`
}

func NewManifest(buffer []byte) (*Manifest, error) {
	manifest := &Manifest{}
	if err := json.Unmarshal(buffer, manifest); err != nil {
		return nil, err
	}
	return manifest, nil
}
