# TODO: Change Copyright to your company if open sourcing or remove header
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

#build stage
ARG BASE=golang:1.16-alpine3.14
FROM ${BASE} AS builder

ARG ALPINE_PKG_BASE="make git gcc libc-dev libsodium-dev zeromq-dev"
ARG ALPINE_PKG_EXTRA=""

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories
RUN apk add --update --no-cache ${ALPINE_PKG_BASE} ${ALPINE_PKG_EXTRA}
WORKDIR /app

COPY . .

RUN [ ! -d "vendor" ] && go mod download all || echo "skipping..."

ARG MAKE="make build"
RUN $MAKE

#final stage
FROM alpine:3.14
# TODO: Change Copyright to your company if open sourcing or remove label
LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2021: Intel'
LABEL Name=new-app-service Version=${VERSION}

# dumb-init is required as security-bootstrapper uses it in the entrypoint script
RUN apk add --update --no-cache ca-certificates zeromq dumb-init

COPY --from=builder /app/Attribution.txt /Attribution.txt
COPY --from=builder /app/LICENSE /LICENSE
COPY --from=builder /app/res/ /res/
COPY --from=builder /app/new-app-service /new-app-service

# TODO: set this port appropriatly as it is in the configuation.toml
EXPOSE 59740

ENTRYPOINT ["/new-app-service"]
CMD ["-cp=consul.http://edgex-core-consul:8500", "--registry", "--confdir=/res"]
