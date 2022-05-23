.PHONY: build release

go.install:
	cd /tmp
	sudo curl -O https://dl.google.com/go/go1.15.linux-amd64.tar.gz
	sudo tar -xf go1.15.linux-amd64.tar.gz
	sudo mv go /usr/local
	cd -

go.get:
	go get ./...

go.fmt:
	go fmt ./...

test:
	gotestsum --format short-verbose --junitfile junit-report.xml --packages="./..." -- -p 1
.PHONY: test

build:
	env GOOS=$(OS) GOARCH=$(ARCH) go build -o artifact
	tar -czvf /tmp/artifact.tar.gz artifact
