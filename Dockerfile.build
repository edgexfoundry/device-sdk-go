#
# Copyright (c) 2019
# Intel
#
# SPDX-License-Identifier: Apache-2.0
#
FROM nexus3.edgexfoundry.org:10004/edgex-golang-base:1.12.6-alpine

COPY --from=docker:latest /usr/local/bin/docker /usr/local/bin/docker

LABEL license='SPDX-License-Identifier: Apache-2.0' \
      copyright='Copyright (c) 2019: Intel'

RUN apk add --update --no-cache zeromq-dev bash

WORKDIR /go/src/github.com/edgexfoundry

COPY go.mod .

RUN go mod download