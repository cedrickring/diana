package registry

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

//For a common registry with basic auth
type V2RegistryClient struct {
	username string
	password string
}

func NewV2RegistryClient(username, password string) Client {
	return &V2RegistryClient{
		username: username,
		password: password,
	}
}

func (v V2RegistryClient) GetManifest(image string) (*Manifest, error) {
	tag, err := name.NewTag(image, name.WeakValidation)
	if err != nil {
		return nil, errors.Wrap(err, "parsing image tag")
	}

	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf(manifestURL, tag.RegistryStr(), tag.RepositoryStr(), tag.TagStr()), nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating manifest request")
	}
	request.Header.Set("Accept", manifestVersionHeader)
	request.SetBasicAuth(v.username, v.password)

	logrus.Infof("Retrieving manifest for image %s", image)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "requesting v2 manifest")
	}
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading manifest body")
	}

	if err := checkResponseCode(response, "failed to get manifest"); err != nil {
		return nil, err
	}

	manifest, err := NewManifest(bytes)
	if err != nil {
		return nil, errors.Wrap(err, "creating new manifest")
	}

	return manifest, nil
}

func (v V2RegistryClient) PullLayer(image string, layer *Layer, out io.Writer) error {
	tag, err := name.NewTag(image, name.WeakValidation)
	if err != nil {
		return errors.Wrap(err, "parsing image tag")
	}

	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf(layerURL, tag.RegistryStr(), tag.RepositoryStr(), layer.Digest), nil)
	if err != nil {
		return errors.Wrap(err, "creating layer pull request")
	}
	request.SetBasicAuth(v.username, v.password)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return errors.Wrapf(err, "requesting layer with sha %s", layer.Digest)
	}
	defer response.Body.Close()

	if err := checkResponseCode(response, "failed to pull layer"); err != nil {
		return err
	}

	length, _ := strconv.Atoi(response.Header.Get("Content-Length"))
	if length != layer.Size {
		return errors.Errorf("invalid content length, expected %d, got %d", layer.Size, length)
	}

	i, err := io.Copy(out, response.Body)
	if err != nil {
		return errors.Wrap(err, "writing layer response to out")
	}

	logrus.Debugf("Bytes written %v", i)

	return nil
}
