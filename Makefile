.PHONY: build release

go.install:
	cd /tmp
	sudo curl -O https://dl.google.com/go/go1.12.linux-amd64.tar.gz
	sudo tar -xf go1.12.linux-amd64.tar.gz
	sudo mv go /usr/local
	cd -

go.get:
	go get ./...

go.fmt:
	go fmt ./...

test:
	go test -v ./...

build:
	env GOOS=$(OS) GOARCH=$(ARCH) go build -o artifact
	tar -czvf /tmp/artifact.tar.gz artifact
