.PHONY: build release

REL_VERSION=$(shell git rev-parse HEAD)
REL_BUCKET=artifacts-play-bucket

install.goreleaser:
	curl -L https://github.com/goreleaser/goreleaser/releases/download/v0.106.0/goreleaser_Linux_x86_64.tar.gz -o /tmp/goreleaser.tar.gz
	tar -xf /tmp/goreleaser.tar.gz -C /tmp
	sudo mv /tmp/goreleaser /usr/bin/goreleaser

go.install:
	cd /tmp
	sudo curl -O https://dl.google.com/go/go1.12.linux-amd64.tar.gz
	sudo tar -xf go1.12.linux-amd64.tar.gz
	sudo mv go /usr/local
	cd -

gsutil.configure:
	gcloud auth activate-service-account gergely-play@artifacts-play.iam.gserviceaccount.com --key-file ~/sec/artifacts-play-84f9d6266402.json
	gcloud config set project artifacts-play

go.get:
	go get

go.fmt:
	go fmt ./...

test:
	go test -v ./...
	# make sure to wipe everything before the test
	go run main.go yank job .
	go run main.go push job -d myTest/myReadme README.md
	go run main.go pull job -d readme2 myTest/myReadme
	go run main.go yank job myTest/myReadme
	diff README.md readme2
	rm readme2

build:
	env GOOS=$(OS) GOARCH=$(ARCH) go build -o artifact
	tar -czvf /tmp/artifact.tar.gz artifact

release:
	$(MAKE) build OS=$(OS) ARCH=$(ARCH) -o artifact
	gsutil cp /tmp/artifact.tar.gz gs://$(REL_BUCKET)/$(REL_VERSION)-$(OS)-$(ARCH).tar.gz
	gsutil acl -R ch -u AllUsers:R gs://$(REL_BUCKET)/$(REL_VERSION)-$(OS)-$(ARCH).tar.gz
	gsutil setmeta -h "Cache-Control:private, max-age=0, no-transform" gs://$(REL_BUCKET)/$(REL_VERSION)-$(OS)-$(ARCH).tar.gz
	echo "https://storage.googleapis.com/$(REL_BUCKET)/$(REL_VERSION)-$(OS)-$(ARCH).tar.gz"
