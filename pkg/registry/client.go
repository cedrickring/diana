package registry

import (
	"io"
	"net/http"

	"github.com/pkg/errors"
)

const (
	manifestURL           = "https://%s/v2/%s/manifests/%s"
	manifestVersionHeader = "application/vnd.docker.distribution.manifest.v2+json"
	layerURL              = "https://%s/v2/%s/blobs/%s"
)

var (
	errAuthRequired = errors.New("authorization required")
)

type Client interface {
	GetManifest(image string) (*Manifest, error)
	PullLayer(image string, layer *Layer, out io.Writer) error
}

func checkResponseCode(r *http.Response, defaultMsg string) error {
	switch r.StatusCode {
	case http.StatusOK:
		return nil
	case http.StatusUnauthorized:
		return errAuthRequired
	case http.StatusNotFound:
		return errors.New("image not found")
	default:
		return errors.New(defaultMsg)
	}
}
