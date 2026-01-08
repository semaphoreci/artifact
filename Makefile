.PHONY: build test

APP_NAME=artifact
MONOREPO_TMP_DIR?=/tmp/monorepo
SECURITY_TOOLBOX_TMP_DIR?=$(MONOREPO_TMP_DIR)/security-toolbox
SECURITY_TOOLBOX_BRANCH ?= main
APP_DIRECTORY ?= /app
SECURITY_SCANNERS=vuln,secret,misconfig

check.prepare:
	rm -rf $(MONOREPO_TMP_DIR)
	git clone --depth 1 --filter=blob:none --sparse https://github.com/semaphoreio/semaphore $(MONOREPO_TMP_DIR) && \
		cd $(MONOREPO_TMP_DIR) && \
		git config core.sparseCheckout true && \
		git sparse-checkout init --cone && \
		git sparse-checkout set security-toolbox && \
		git checkout main && cd -

check.static: check.prepare
	docker run -it -v $$(pwd):$(APP_DIRECTORY) \
		-v $(SECURITY_TOOLBOX_TMP_DIR):$(SECURITY_TOOLBOX_TMP_DIR) \
		registry.semaphoreci.com/ruby:3 \
		bash -c 'cd $(APP_DIRECTORY) && $(SECURITY_TOOLBOX_TMP_DIR)/code --language go -d'

check.deps: check.prepare
	docker run -it -v $$(pwd):$(APP_DIRECTORY) \
		-v $(SECURITY_TOOLBOX_TMP_DIR):$(SECURITY_TOOLBOX_TMP_DIR) \
		registry.semaphoreci.com/ruby:3 \
		bash -c 'cd $(APP_DIRECTORY) && $(SECURITY_TOOLBOX_TMP_DIR)/dependencies --language go -d'

check.generate-report: check.prepare
	docker run -it \
		-v $$(pwd):/app \
		-v $(SECURITY_TOOLBOX_TMP_DIR):$(SECURITY_TOOLBOX_TMP_DIR) \
		registry.semaphoreci.com/ruby:3 \
		bash -c 'cd $(APP_DIRECTORY) && $(SECURITY_TOOLBOX_TMP_DIR)/report --service-name "[$(CHECK_TYPE)] $(APP_NAME)"'

check.generate-global-report: check.prepare
	docker run -it \
		-v $$(pwd):/app \
		-v $(SECURITY_TOOLBOX_TMP_DIR):$(SECURITY_TOOLBOX_TMP_DIR) \
		registry.semaphoreci.com/ruby:3 \
		bash -c 'cd $(APP_DIRECTORY) && $(SECURITY_TOOLBOX_TMP_DIR)/global-report -i reports -o out'

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
