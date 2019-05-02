package docker

import (
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cedrickring/diana/pkg/util"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/pkg/errors"
)

type dockerConfig struct {
	Auths map[string]auth `json:"auths"`
}

type auth struct {
	Auth string `json:"auth"`
}

func GetCredentials(registry string) (username, password string, err error) {
	home := util.HomeDir()
	if home == "" {
		return "", "", errors.New("Can't find docker config at ~/.docker/config.json")
	}

	dockerConfigPath := filepath.Join(home, ".docker", "config.json")
	if _, err := os.Stat(dockerConfigPath); os.IsNotExist(err) {
		return "", "", errors.New("Can't find docker config at ~/.docker/config.json")
	}

	bytes, err := ioutil.ReadFile(dockerConfigPath)
	if err != nil {
		return "", "", errors.Wrap(err, "reading docker config")
	}

	cfg := dockerConfig{}
	if err := json.Unmarshal(bytes, &cfg); err != nil {
		return "", "", errors.Wrap(err, "unmarshaling docker config")
	}

	if registry == name.DefaultRegistry {
		registry = "https://index.docker.io/v1/" // the docker hub credentials are stored with the full url
	}

	auth, ok := cfg.Auths[registry]
	if !ok {
		return "", "", errors.Errorf("Couldn't find credentials for registry %s", registry)
	}

	b64, err := base64.StdEncoding.DecodeString(auth.Auth)
	if err != nil {
		return "", "", errors.Wrap(err, "decoding auth string")
	}

	split := strings.Split(string(b64), ":")

	return split[0], split[1], nil
}
