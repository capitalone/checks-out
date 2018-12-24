# Propagate git tag to version identifier
VERSION_TAG:=$(shell git describe --abbrev=0 --tags)
ifeq ($(VERSION_TAG),)
VERSION_TAG:=noversion
endif
# Propagate git revision number to version identifier
VERSION:=${VERSION_TAG}+$(shell git rev-parse --short HEAD)

# directories and remove vendor directory
# used for running unit tests
NOVENDOR := $(shell go list -e ./... | grep -v /vendor/)

# files and remove vendor directory, auto generated files, and mock files
# used for static analysis, code linting, and code formatting
NOVENDOR_FILES := $(shell find . -name "*.go" | grep -v /vendor/ | grep -v /mock/ | grep -v "_gen\.go" )

all: build

gen: gen-assets gen-migration

gen-assets:
	go generate github.com/capitalone/checks-out/web/static
	go generate github.com/capitalone/checks-out/web/template

gen-migration:
	go generate github.com/capitalone/checks-out/store/migration

build:
	@# static linking sqlite seems to break DNS
	go build --ldflags '-X github.com/capitalone/checks-out/version.Version=$(VERSION)' -o checks-out
	go test -run ^$$ $(NOVENDOR) > /dev/null

test: vet
	@# use redirect instead of tee to preserve exit code
	go test -short -p=1 -cover $(NOVENDOR) -v > report.out; \
	code=$$?; \
	cat report.out; \
	grep -e 'FAIL' report.out; \
	exit $${code}

test-complete: vet
	@# use redirect instead of tee to preserve exit code
	GITHUB_TEST_ENABLE=1 go test -short -p=1 -cover $(NOVENDOR) -v > report.out; \
	code=$$?; \
	cat report.out; \
	grep -e 'FAIL' report.out; \
	exit $${code}

.PHONY: test-cover-html

test-cover-html:
	echo "mode: count" > coverage-all.out
	echo "Packages: $(NOVENDOR)"
	echo "VERSION: $(VERSION)"
	$(foreach pkg,$(NOVENDOR), \
		echo ${pkg}; \
		go test -coverprofile=coverage.out -covermode=count ${pkg}; \
		tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out -o coverage.html

fmt:
	for FILE in $(NOVENDOR_FILES); do go fmt $$FILE; done;

vet:
	@echo 'Running go tool vet -shadow'
	@for FILE in $(NOVENDOR_FILES); do go tool vet -shadow $$FILE || exit 1; done;

lint:
	for FILE in $(NOVENDOR_FILES); do golint $$FILE; done;

clean:
	rm -f checks-out report.out

test-mysql:
	DB_DRIVER="mysql" DB_SOURCE="root@tcp(127.0.0.1:3306)/test?parseTime=true" go test -v -cover github.com/capitalone/checks-out/store/datastore
