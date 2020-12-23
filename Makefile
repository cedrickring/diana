all: lint test build

format: vet fmt

build:
	mkdir -p bin
	go get
	go build -o bin/diana

fmt:
	go fmt ./pkg/... ./cmd/...

vet:
	go vet ./pkg/... ./cmd/...

imports:
	go get golang.org/x/tools/cmd/goimports
	goimports -w ./pkg/* ./cmd/*

lint: require-gopath
	bash scripts/install_golangci-lint.sh
	golangci-lint run --no-config \
    	-E goconst \
    	-E goimports \
    	-E gocritic \
    	-E golint \
    	-E interfacer \
    	-E maligned \
    	-E misspell \
    	-E unconvert \
    	-E unparam \
    	-D errcheck \
      --skip-dirs vendor

gox:
	go get github.com/mitchellh/gox

test:
	go test ./...

install: all
	sudo cp bin/kbuild /usr/local/bin/kbuild-dev

build-all:
	mkdir -p out && cd out
	which gox || make gox
	gox -arch="amd64" -os="darwin linux windows" --output "out/diana_{{.OS}}_{{.Arch}}" github.com/cedrickring/diana/cmd

require-gopath:
ifndef GOPATH
  $(error GOPATH is not set)
endif