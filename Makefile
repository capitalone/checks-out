# Propagate git tag to version identifier
VERSION_TAG:=$(shell git describe --abbrev=0 --tags)
ifeq ($(VERSION_TAG),)
VERSION_TAG:=noversion
endif
# Propagate git revision number to version identifier
VERSION:=${VERSION_TAG}+$(shell git rev-parse --short HEAD)

all: build

gen: gen-assets gen-migration

gen-assets:
	go generate github.com/capitalone/checks-out/web/static
	go generate github.com/capitalone/checks-out/web/template

gen-migration:
	go generate github.com/capitalone/checks-out/store/migration

build: test
	@# static linking sqlite seems to break DNS
	go build --ldflags '-X github.com/capitalone/checks-out/version.Version=$(VERSION)' -o checks-out

test: vet
	@# use redirect instead of tee to preserve exit code
	go test -short -cover -v ./... > report.out; \
	code=$$?; \
	cat report.out; \
	grep -e 'FAIL' report.out; \
	exit $${code}
.PHONY: test

test-complete: vet
	@# use redirect instead of tee to preserve exit code
	GITHUB_TEST_ENABLE=1 go test -short -cover -v ./... > report.out; \
	code=$$?; \
	cat report.out; \
	grep -e 'FAIL' report.out; \
	exit $${code}

.PHONY: test-cover-html

test-cover-html:
	@echo "VERSION: $(VERSION)"
	go test -coverprofile=coverage.out -covermode=count ./...
	go tool cover -html=coverage.out -o coverage.html

fmt:
	go fmt ./...

vet: get-modules
	@echo 'Running go tool vet -shadow'
	go vet ./...

lint: get-modules
	golint ./...

clean:
	rm -f checks-out report.out

get-modules:
	go mod download
	go mod verify

test-mysql:
	DB_DRIVER="mysql" DB_SOURCE="root@tcp(127.0.0.1:3306)/test?parseTime=true" go test -v -cover github.com/capitalone/checks-out/store/datastore
