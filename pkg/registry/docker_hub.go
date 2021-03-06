package registry

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	bearerAuthURL        = "https://auth.docker.io/token?service=registry.docker.io&scope=repository:%s:pull"
	jwtAuthURL           = "https://hub.docker.com/v2/users/login/"
	hubManifest          = "https://"
	bearerAuth    string = "Bearer"
)

type tokenResponse struct {
	Token string `json:"token"`
}

type jwtRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type DockerHubRegistryClient struct {
	username string
	password string
}

func NewDockerHubRegistryClient(username, password string) Client {
	return &DockerHubRegistryClient{
		username: username,
		password: password,
	}
}

func (d DockerHubRegistryClient) getBearerToken(repository string) (string, error) {
	var resp *http.Response
	var err error

	resp, err = http.Get(fmt.Sprintf(bearerAuthURL, repository))
	if err != nil {
		return "", errors.Wrap(err, "getting bearer token")
	}
	defer resp.Body.Close()

	if err := checkResponseCode(resp, "failed to get bearer token"); err != nil {
		if err == errAuthRequired { //try basic auth
			req := jwtRequest{
				Username: d.username,
				Password: d.password,
			}

			buf := &bytes.Buffer{}
			if err := json.NewEncoder(buf).Encode(req); err != nil {
				return "", err
			}

			resp, err = http.Post(jwtAuthURL, "application/json", buf)
			if err != nil {
				return "", errors.Wrap(err, "getting jwt token from docker hub")
			}

			if err := checkResponseCode(resp, "failed to get bearer token"); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}

	token := tokenResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return "", errors.Wrap(err, "decoding json response")
	}

	return token.Token, nil
}

func (d DockerHubRegistryClient) GetManifest(image string) (*Manifest, error) {
	tag, err := name.NewTag(image, name.WeakValidation)
	if err != nil {
		return nil, errors.Wrap(err, "parsing image tag")
	}

	bearer, err := d.getBearerToken(tag.RepositoryStr())
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf(manifestURL, tag.RegistryStr(), tag.RepositoryStr(), tag.TagStr()), nil)
	if err != nil {
		return nil, errors.Wrap(err, "creating manifest request")
	}
	request.Header.Set("Accept", manifestVersionHeader)
	request.Header.Set("Authorization", fmt.Sprintf("%s %s", bearerAuth, bearer))

	logrus.Infof("Retrieving manifest for image %s", image)

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, "requesting v2 manifest")
	}
	defer response.Body.Close()

	if err := checkResponseCode(response, "failed to get manifest"); err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "reading manifest body")
	}

	manifest, err := NewManifest(b)
	if err != nil {
		return nil, errors.Wrap(err, "creating new manifest")
	}

	return manifest, nil
}

func (d DockerHubRegistryClient) PullLayer(image string, layer *Layer, out io.Writer) error {
	tag, err := name.NewTag(image, name.WeakValidation)
	if err != nil {
		return errors.Wrap(err, "parsing image tag")
	}

	bearer, err := d.getBearerToken(tag.RepositoryStr())
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf(layerURL, tag.RegistryStr(), tag.RepositoryStr(), layer.Digest), nil)
	if err != nil {
		return errors.Wrap(err, "creating layer pull request")
	}
	request.Header.Set("Authorization", fmt.Sprintf("%s %s", bearerAuth, bearer))

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return errors.Wrapf(err, "requesting layer with sha %s", layer.Digest)
	}
	defer response.Body.Close()

	if err := checkResponseCode(response, "failed to get layer"); err != nil {
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
