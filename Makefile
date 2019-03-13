.PHONY: build release

REL_VERSION=$(shell git rev-parse HEAD)
REL_BUCKET=artifact-cli-releases

go.install:
	cd /tmp
	sudo curl -O https://dl.google.com/go/go1.11.linux-amd64.tar.gz
	sudo tar -xf go1.11.linux-amd64.tar.gz
	sudo mv go /usr/local
	cd -

#gsutil.configure:
#	gcloud auth activate-service-account deploy-from-semaphore@semaphore2-prod.iam.gserviceaccount.com --key-file ~/gce-creds.json
#	gcloud config set project semaphore2-prod

go.get:
	go get

go.fmt:
	go fmt ./...

test:
	go test ./...
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
#	gsutil cp /tmp/artifact.tar.gz gs://$(REL_BUCKET)/$(REL_VERSION)-$(OS)-$(ARCH).tar.gz
#	gsutil acl -R ch -u AllUsers:R gs://$(REL_BUCKET)/$(REL_VERSION)-$(OS)-$(ARCH).tar.gz
#	gsutil setmeta -h "Cache-Control:private, max-age=0, no-transform" gs://$(REL_BUCKET)/$(REL_VERSION)-$(OS)-$(ARCH).tar.gz
#	echo "https://storage.googleapis.com/$(REL_BUCKET)/$(REL_VERSION)-$(OS)-$(ARCH).tar.gz"

release.all:
	$(MAKE) release OS=linux   ARCH=386
	$(MAKE) release OS=linux   ARCH=amd64
	$(MAKE) release OS=darwin  ARCH=386
	$(MAKE) release OS=darwin  ARCH=amd64
	# $(MAKE) release OS=windows ARCH=386    # mousetrap issues?
	# $(MAKE) release OS=windows ARCH=amd64

release.stable:
	$(MAKE) release.all REL_VERSION=stable

release.edge:
	$(MAKE) release.all REL_VERSION=edge

release.stable.install.script:
#	gsutil cp scripts/get gs://$(REL_BUCKET)/get.sh
#	gsutil acl -R ch -u AllUsers:R gs://$(REL_BUCKET)/get.sh
#	gsutil setmeta -h "Cache-Control:private, max-age=0, no-transform" gs://$(REL_BUCKET)/get.sh
#	echo "https://storage.googleapis.com/$(REL_BUCKET)/get.sh"

release.edge.install.script:
#	gsutil cp scripts/get-edge gs://$(REL_BUCKET)/get-edge.sh
#	gsutil acl -R ch -u AllUsers:R gs://$(REL_BUCKET)/get-edge.sh
#	gsutil setmeta -h "Cache-Control:private, max-age=0, no-transform" gs://$(REL_BUCKET)/get-edge.sh
#	echo "https://storage.googleapis.com/$(REL_BUCKET)/get-edge.sh"
