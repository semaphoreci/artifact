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
