.PHONY: build test clean prepare update

GO=CGO_ENABLED=0 go
GOFLAGS=-ldflags

MICROSERVICES=examples/simple/cmd/simple-device
.PHONY: $(MICROSERVICES)

build: $(MICROSERVICES)
	go build ./...

examples/simple/cmd/simple-device:
	$(GO) build -o $@ ./examples/simple/cmd/ 

test:
	go test ./... -cover

clean:
	rm -f $(MICROSERVICES)

prepare:
	glide install

update:
	glide update
