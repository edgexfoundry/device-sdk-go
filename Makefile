.PHONY: build test clean prepare update

GO=CGO_ENABLED=0 go

MICROSERVICES=example/cmd/device-simple/device-simple
.PHONY: $(MICROSERVICES)

VERSION=$(shell cat ./VERSION)

GOFLAGS=-ldflags "-X github.com/edgexfoundry/device-sdk-go.Version=$(VERSION)"

build: $(MICROSERVICES)
	go build ./...

example/cmd/device-simple/device-simple:
	$(GO) build $(GOFLAGS) -o $@ ./example/cmd/device-simple

test:
	go test ./... -cover

clean:
	rm -f $(MICROSERVICES)

prepare:
	glide install

update:
	glide update
