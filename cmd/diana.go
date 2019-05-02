package main

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cedrickring/diana/pkg/docker"
	"github.com/cedrickring/diana/pkg/registry"
	"github.com/cedrickring/diana/pkg/tar"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	image            string
	includeBaseLayer bool
	forceTTYColors   bool
)

func main() {

	rootCmd := cobra.Command{
		Use: "diana",
		Run: runCommand,
	}

	rootCmd.Flags().StringVarP(&image, "image", "i", "", "Full image name")
	rootCmd.Flags().BoolVarP(&includeBaseLayer, "base-layer", "", false, "Specify to also pull the base image layer")
	rootCmd.Flags().BoolVarP(&forceTTYColors, "color", "c", false, "Force logrus coloful output")
	rootCmd.MarkFlagRequired("image")

	rootCmd.Execute()
}

func runCommand(_ *cobra.Command, args []string) {
	setupLogrus()

	if len(args) == 0 {
		logrus.Fatalf("Please specify the file to be extracted as the first argument")
	}
	fileName := args[0]

	tag, err := name.NewTag(image, name.WeakValidation)
	if err != nil {
		logrus.Fatal(err)
	}

	var username, password string
	if !strings.HasPrefix(tag.RepositoryStr(), "library") {
		username, password, err = docker.GetCredentials(tag.RegistryStr())
		if err != nil {
			logrus.WithError(err).Fatalf("Couldn't find registry credentials")
		}
	}

	var client registry.Client
	if tag.RegistryStr() == name.DefaultRegistry { //dockerhub
		client = registry.NewDockerHubRegistryClient(username, password)
	} else {
		client = registry.NewV2RegistryClient(username, password)
	}

	manifest, err := client.GetManifest(image)
	if err != nil {
		logrus.WithError(err).Fatalf("Failed to get manifest for image %s", image)
	}

	layers := manifest.Layers
	if !includeBaseLayer {
		//drop the first layer as it's probably the base image
		layers = layers[1:]
	}

	tmp, err := ioutil.TempDir("", "diana")
	if err != nil {
		logrus.WithError(err).Errorf("Can't create temporary directory. Please check rights for this executable.")
		return
	}

	var tarFiles []string
	for _, layer := range layers {
		logrus.Infof("Pulling layer %s (%d B)", layer.Digest, layer.Size)

		f, _ := ioutil.TempFile(tmp, "*.tar.gz")
		err := client.PullLayer(image, &layer, f)
		if err != nil {
			logrus.Errorln(err)
			f.Close()
			return
		}

		tarFiles = append(tarFiles, f.Name())
		f.Close()
	}
	defer func() {
		os.RemoveAll(tmp)
	}()

	path := filepath.Join(tmp, "image")
	for _, f := range tarFiles {
		tar.UnTar(path, f)
	}

	search := filepath.Join(path, fileName)
	f, err := os.Open(search)
	if err != nil {
		logrus.Errorf(`The file "%v" doesn't exist in the image`, fileName)
		return
	}
	defer f.Close()

	fileName = filepath.ToSlash(fileName)
	if strings.Contains(fileName, "/") {
		fileName = fileName[strings.LastIndex(fileName, "/")+1:]
	}

	target, err := os.OpenFile(fileName, os.O_CREATE|os.O_RDWR, os.ModePerm)
	if err != nil {
		logrus.WithError(err).Errorf("Couldn't create target file")
		return
	}
	defer target.Close()

	if _, err = io.Copy(target, f); err != nil {
		logrus.Fatal(err)
	}

	logrus.Infof("Extracted file to ./%s", fileName)
}

func setupLogrus() {
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors: forceTTYColors,
	})
	logrus.SetOutput(os.Stdout)
}
