.PHONY: build test clean docker

GO=CGO_ENABLED=1 GO111MODULE=on go

MICROSERVICES=examples/simple-filter-xml/simple-filter-xml examples/simple-filter-xml-mqtt/simple-filter-xml-mqtt examples/simple-filter-xml-post/simple-filter-xml-post
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X app-functions-sdk-go.Version=$(VERSION)"

GIT_SHA=$(shell git rev-parse HEAD)

build: $(MICROSERVICES)
	$(GO) build ./...

examples/simple-filter-xml/simple-filter-xml:
	$(GO) build $(GOFLAGS) -o $@ ./examples/simple-filter-xml

examples/simple-filter-xml-mqtt/simple-filter-xml-mqtt:
	$(GO) build $(GOFLAGS) -o $@ ./examples/simple-filter-xml-mqtt

examples/simple-filter-xml-post/simple-filter-xml-post:
	$(GO) build $(GOFLAGS) -o $@ ./examples/simple-filter-xml-post

docker:
	docker build \
		-f examples/simple-filter-xml/Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/docker-app-functions-sdk-go-simple:$(GIT_SHA) \
		-t edgexfoundry/docker-app-functions-sdk-go-simple:$(VERSION)-dev \
		.

test:
	$(GO) test ./... -cover
	$(GO) vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]

clean:
	rm -f $(MICROSERVICES)