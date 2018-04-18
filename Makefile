BUILD_DATA = $(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
PROJECT_NAME = $(shell basename ${PWD})
PROJECT_DIR = $(shell ${PWD})
COMMIT_SHA = $(shell git rev-parse --short HEAD || echo "GitNotFound")
MODULE_NAME = $(shell go list)
VERSION = $(shell git describe --tags --exact-match 2> /dev/null || git symbolic-ref -q --short HEAD)
LDFLAGS = \
	-X main.name=$(PROJECT_NAME) \
	-X main.date=$(BUILD_DATA) \
	-X main.commit=$(COMMIT_SHA) \
	-X main.version=$(VERSION)

.PHONY: all
all: depslint lint deps test cross ## execute all commands

.PHONY: deps
deps: ## install dependencies
	go get -u github.com/golang/dep/cmd/dep github.com/mitchellh/gox
	dep ensure

.PHONY: binary
binary: ## build binary to current OS
	gox -output "$(PROJECT_NAME)" \
		-os `go env GOHOSTOS` \
		-arch=`go env GOHOSTARCH` \
		-ldflags='$(LDFLAGS)' $(MODULE_NAME)

.PHONY: cross
cross: ## build binary cross OS
	rm -rf ./build/$(PROJECT_NAME)*
	gox -output "./build/{{.Dir}}-{{.OS}}-{{.Arch}}" \
		-os "linux darwin windows" \
		-arch="amd64" \
		-ldflags="$(LDFLAGS)" $(MODULE_NAME)

.PHONY: test
test: ## test go files
	go test -cover $(shell go list ./...|grep -v '/vendor/')

.PHONY: depslint
depslint: ## install lint dependencies: gofmt, govet, golint, gocyclo, ineffassign
	go get -u gopkg.in/alecthomas/gometalinter.v2
	go get -u github.com/golang/lint/golint
	go get -u github.com/dnephin/govet
	go get -u github.com/alecthomas/gocyclo
	go get -u github.com/gordonklaus/ineffassign
	go get -u honnef.co/go/tools/cmd/gosimple
	go get -u github.com/walle/lll/...
	go get -u github.com/mdempsky/unconvert
	go get -u honnef.co/go/tools/cmd/unused
	go get -u github.com/tsenart/deadcode
	go get -u golang.org/x/tools/cmd/goimports
	go get -u mvdan.cc/interfacer
	go get -u mvdan.cc/unparam
	go get -u github.com/client9/misspell/cmd/misspell
	go get -u github.com/alexkohler/nakedret

.PHONY: lint
lint: ## run all the lint tools (see gometalinter.json)
	gometalinter.v2 --config gometalinter.json --vendor ./...

.PHONY: help
help: ## print this help
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {sub("\\\\n",sprintf("\n%22c"," "), $$2);printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)
