.PHONY: build test

SECURITY_TOOLBOX_BRANCH ?= master
SECURITY_TOOLBOX_TMP_DIR ?= /tmp/security-toolbox

check.prepare:
	rm -rf $(SECURITY_TOOLBOX_TMP_DIR)
	git clone git@github.com:renderedtext/security-toolbox.git $(SECURITY_TOOLBOX_TMP_DIR) && (cd $(SECURITY_TOOLBOX_TMP_DIR) && git checkout $(SECURITY_TOOLBOX_BRANCH) && cd -)

check.static: check.prepare
	docker run -it -v $$(pwd):/app \
		-v $(SECURITY_TOOLBOX_TMP_DIR):$(SECURITY_TOOLBOX_TMP_DIR) \
		registry.semaphoreci.com/ruby:2.7 \
		bash -c 'cd /app && $(SECURITY_TOOLBOX_TMP_DIR)/code --language go -d'

check.deps: check.prepare
	docker run -it -v $$(pwd):/app \
		-v $(SECURITY_TOOLBOX_TMP_DIR):$(SECURITY_TOOLBOX_TMP_DIR) \
		registry.semaphoreci.com/ruby:2.7 \
		bash -c 'cd /app && $(SECURITY_TOOLBOX_TMP_DIR)/dependencies --language go -d'

go.get:
	docker-compose run --rm cli go get ./...

go.fmt:
	docker-compose run --rm cli go fmt ./...

test:
	docker-compose run --rm cli gotestsum --format short-verbose --junitfile junit-report.xml --packages="./..." -- -p 1

# Go 1.20 changed the handling of git worktrees,
# so we need to pass buildvcs=false, for now.
# See: https://github.com/golang/go/issues/59068
build:
	docker-compose run --rm cli env GOFLAGS=-buildvcs=false GOOS=$(OS) GOARCH=$(ARCH) go build -o artifact
