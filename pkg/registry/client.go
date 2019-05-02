package registry

import "io"

const (
	manifestURL           = "https://%s/v2/%s/manifests/%s"
	manifestVersionHeader = "application/vnd.docker.distribution.manifest.v2+json"
	layerURL              = "https://%s/v2/%s/blobs/%s"
)

type Client interface {
	GetManifest(image string) (*Manifest, error)
	PullLayer(image string, layer *Layer, out io.Writer) error
}
