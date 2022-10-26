BUILD_DIR	?=	build
AUTHOR_FILE	?=	author

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

##    TWTD    ##
TWTD_VERSION	?=	$(shell cat VERSION)+$(shell cat $(AUTHOR_FILE))-$(shell date +"%Y%m%d")
TWTD_FLAGS		?=	-dir public
TWTD_USR		?=	user
TWTD_PWD		?=	Password1!
TWTD_GO_FILES  	:=	$(shell go list -f '{{ $$dir := .Dir }}{{ range .GoFiles }}{{ printf "%s/%s " $$dir . }}{{ end }}' -deps ./cmd/twtd/... | grep $$(pwd))

.PHONY: run/twtd
run/twtd:
	TWTD_USR=$(TWTD_USR) TWTD_PWD=$(TWTD_PWD) go run ./cmd/twtd/... $(TWTD_FLAGS)

.PHONY: build/twtd
build/twtd: $(BUILD_DIR)/$(shell go env GOOS)/$(shell go env GOARCH)/twtd

.PHONY: container/twtd
container/twtd: $(BUILD_DIR)/linux/arm64/twtd $(BUILD_DIR)/linux/amd64/twtd
	@(docker buildx create --name twtd_builder --node twtd_builder &> /dev/null || true)
	docker buildx build \
		--builder twtd_builder \
		--platform=linux/amd64,linux/arm64 \
		--push \
		--label "git-commit=$(shell git rev-parse HEAD)" \
		--label "version=$(TWTD_VERSION)" \
		-t m25n/twtd:$(subst +,-,$(TWTD_VERSION)) \
		-f Dockerfile \
		$(BUILD_DIR)

$(BUILD_DIR)/%/twtd: go.mod go.sum $(TWTD_GO_FILES)
	GOOS=$(shell basename $$(dirname $$(dirname $@))) \
	GOARCH=$(shell basename $$(dirname $@)) \
	CGO_ENABLED=0 \
	go build \
		-ldflags "-X main.version=$(TWTD_VERSION) -X main.gitCommit=$(shell git rev-parse HEAD)" \
		-o $@ ./cmd/twtd/...


