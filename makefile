.PHONY: build test clean docker

GO=CGO_ENABLED=0 GO111MODULE=on go

MICROSERVICES=examples/simple-filter-xml/simple-filter-xml
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X github.com/edgexfoundry/app-functions-sdk-go.Version=$(VERSION)"

GIT_SHA=$(shell git rev-parse HEAD)

build: $(MICROSERVICES)
	$(GO) build ./...

examples/simple-filter-xml/simple-filter-xml:
	$(GO) build $(GOFLAGS) -o $@ ./examples/simple-filter-xml

docker:
	docker build \
		-f examples/simple-filter-xml/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-app-functions-sdk-go-simple:$(GIT_SHA) \
		-t edgexfoundry/docker-app-functions-sdk-go-simple:$(VERSION)-dev \
		.

test:
	$(GO) test ./... -cover

clean:
	rm -f $(MICROSERVICES)