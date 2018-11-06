.PHONY: build test clean prepare update

GO=CGO_ENABLED=0 go
GOFLAGS=-ldflags

MICROSERVICES=example/cmd/device-simple/device-simple
.PHONY: $(MICROSERVICES)

build: $(MICROSERVICES)
	go build ./...

example/cmd/device-simple/device-simple:
	$(GO) build -o $@ ./example/cmd/device-simple

test:
	go test ./... -cover

clean:
	rm -f $(MICROSERVICES)

prepare:
	glide install

update:
	glide update
