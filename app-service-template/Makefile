#
# Copyright (c) 2021 Intel Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

.PHONY: build test clean docker

GO=CGO_ENABLED=1 go

# VERSION file is not needed for local development, In the CI/CD pipeline, a temporary VERSION file is written
# if you need a specific version, just override below
# TODO: If your service is not being upstreamed to Edgex Foundry, you need to determine the best approach for
#       setting your service's version for non-development builds.
APPVERSION=$(shell cat ./VERSION 2>/dev/null || echo 0.0.0)

# This pulls the version of the SDK from the go.mod file. If the SDK is the only required module,
# it must first remove the word 'required' so the offset of $2 is the same if there are multiple required modules
SDKVERSION=$(shell cat ./go.mod | grep 'github.com/edgexfoundry/app-functions-sdk-go/v2 v' | sed 's/require//g' | awk '{print $$2}')

MICROSERVICE=new-app-service
GOFLAGS=-ldflags "-X github.com/edgexfoundry/app-functions-sdk-go/v2/internal.SDKVersion=$(SDKVERSION) -X github.com/edgexfoundry/app-functions-sdk-go/v2/internal.ApplicationVersion=$(APPVERSION)"

# TODO: uncomment and remove default once files are in a Github repository or
#       remove totally including usage below
#GIT_SHA=$(shell git rev-parse HEAD)
GIT_SHA=no-sha

build:
	$(GO) build $(GOFLAGS) -o $(MICROSERVICE)

# TODO: Change the registries (edgexfoundry, nexus3.edgexfoundry.org:10004) below as needed.
#       Leave them as is if service is to be upstreamed to EdgeX Foundry
# NOTE: This is only used for local development. Jenkins CI does not use this make target
docker:
	docker build \
	    --build-arg http_proxy \
	    --build-arg https_proxy \
		-f Dockerfile \
		--label "git_sha=$(GIT_SHA)" \
		-t edgexfoundry/new-app-service:$(GIT_SHA) \
		-t edgexfoundry/new-app-service:master-dev \
		-t nexus3.edgexfoundry.org:10004/new-app-service:master-dev \
		.

# The test-go-mod-tidy.sh & test-attribution-txt.sh scripts are required for upstreaming to EdgeX Foundry.
# TODO: Remove bin folder and reference to scripts below if NOT upstreaming to EdgeX Foundry.
test:
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) vet ./...
	gofmt -l .
	[ "`gofmt -l .`" = "" ]
	./bin/test-go-mod-tidy.sh
	./bin/test-attribution-txt.sh

clean:
	rm -f $(MICROSERVICE)


