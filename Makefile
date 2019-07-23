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
	SEMAPHORE_ARTIFACT_TOKEN=localhost:8080 SEMAPHORE_PROJECT_ID=some_project SEMAPHORE_WORKFLOW_ID=some_workflow SEMAPHORE_JOB_ID=some_job go test -v ./...

build:
	env GOOS=$(OS) GOARCH=$(ARCH) go build -o artifact
	tar -czvf /tmp/artifact.tar.gz artifact
