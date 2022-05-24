.PHONY: build release test

go.get:
	go get ./...

go.fmt:
	go fmt ./...

test:
	gotestsum --format short-verbose --junitfile junit-report.xml --packages="./..." -- -p 1

build:
	env GOOS=$(OS) GOARCH=$(ARCH) go build -o artifact
