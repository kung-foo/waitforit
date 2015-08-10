VERSION := v0.0.1
BUILDSTRING := $(shell git log --pretty=format:'%h' -n 1)
VERSIONSTRING := waitforit version $(VERSION)+$(BUILDSTRING)

ifndef GOARCH
	GOARCH := $(shell go env GOARCH)
endif

ifndef GOOS
	GOOS := $(shell go env GOOS)
endif

OUTPUT := dist/$(GOARCH)/$(GOOS)/waitforit-$(GOOS)-$(GOARCH)

ifeq ($(GOOS), windows)
	OUTPUT := $(OUTPUT).exe
endif

.PHONY: default build clean update-godeps test

default: build

build: $(OUTPUT)

$(OUTPUT): app/main.go waitforit.go redis.go rdbms.go
	mkdir -p dist/$(GOARCH)/$(GOOS)
	CGO_ENABLED=0 godep go build -v -o $(OUTPUT) -ldflags "-X main.VERSION \"$(VERSIONSTRING)\"" app/main.go

clean:
	rm -rf dist

update-godeps:
	rm -rf Godeps
	godep save

test:
	godep go test -cover -v ./...
