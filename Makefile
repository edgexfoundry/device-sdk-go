.PHONY: test

GO=CGO_ENABLED=1 GO111MODULE=on go

build:
	make -C ./app-service-template build

test-template:
	make -C ./app-service-template test

test: build test-template
	$(GO) test ./... -coverprofile=coverage.out ./...
	$(GO) vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]
	./app-service-template/bin/test-go-mod-tidy.sh