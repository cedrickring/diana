# diana

A simple executable to extract binaries out of your container images.

### Usage

`diana -i <image> /path/to/binary`

E.g. to extract the `helloworld` binary out of my `cedrickring/hello-world` image, just type
```bash
./diana -i cedrickring/hello-world /app/helloworld
```
and the helloworld binary will be extracted to ./helloworld

Example output:
```
INFO[0000] Retrieving manifest for image cedrickring/hello-world 
INFO[0001] Pulling layer sha256:20e3fdddeb7723652929c6d14ec952eb3810069340118027bddfebe765f1eaf4 (965528 B) 
INFO[0002] Extracted file to ./helloworld 
```

**Note**: to pull private images, just have a `~/.docker/config.json` in place with your credentials.

### Installation

```
curl -fsSL https://raw.githubusercontent.com/cedrickring/diana/master/scripts/get | bash
```
or download the binary manually from the [releases tab](https://github.com/cedrickring/diana/releases) and put it in
your `$PATH`.

(No docker is required)

### Flags

- `-i/--image` The image containing the file to be extracted
- `--base-layer` Pull the base image layer too (if you want to extract a file from a base image) 
- `-c/--color` Force colorful terminal output

### Why use diana instead of just `docker cp` ???

Well with `diana` you're not pulling the base image layer, but all the other layers which might contain the
binary. So the download time will be way faster than pulling the whole image down from e.g. Docker Hub.

Aaaaand `diana` will cleanup after she's done (instead of letting you sit on GBs of images) ;)

#### Notes

The name `diana` comes from a (imo) very cool [keynote](https://www.youtube.com/watch?v=oNa3xK2GFKY) by [@kelseyhightower](https://github.com/kelseyhightower) at
KubeCon 2018 where he extracts binaries out of an image with a tool called diana. Unfortunately he didn't make the project
available to all of us so I decided to make it by myself.
