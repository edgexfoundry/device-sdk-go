#
# Copyright (c) 2019 Intel Corporation
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

FROM golang:1.11-alpine AS builder

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2019: Intel'

# add git for go modules
RUN apk update && apk add --no-cache make git gcc libc-dev libsodium-dev zeromq-dev
ENV GO111MODULE=on
WORKDIR /go/src/github.com/edgexfoundry/app-functions-sdk-go

COPY go.mod .

RUN go mod download

COPY . .

RUN make examples/simple-filter-xml-post/simple-filter-xml-post

# Next image - Copy built Go binary into new workspace
FROM alpine

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2019: Intel'

RUN apk --no-cache add zeromq
COPY --from=builder /go/src/github.com/edgexfoundry/app-functions-sdk-go/examples/simple-filter-xml-post/res/docker /res/docker
COPY --from=builder /go/src/github.com/edgexfoundry/app-functions-sdk-go/examples/simple-filter-xml-post/simple-filter-xml-post /simple-filter-xml-post

CMD [ "/simple-filter-xml-post","--registry","--profile=docker","--confdir=/res"]