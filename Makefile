.PHONY: build release

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
