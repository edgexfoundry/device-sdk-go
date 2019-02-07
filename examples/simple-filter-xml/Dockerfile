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

FROM golang:1.11.2-alpine3.7 AS builder

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2019: Intel'

# The main mirrors are giving us timeout issues on builds periodically.
# So we can try these.
RUN echo http://nl.alpinelinux.org/alpine/v3.7/main > /etc/apk/repositories; \
  echo http://nl.alpinelinux.org/alpine/v3.7/community >> /etc/apk/repositories

# add git for go modules
RUN apk update && apk add --no-cache make git gcc libc-dev

WORKDIR /github.com/edgexfoundry/app-functions-sdk-go

COPY . /github.com/edgexfoundry/app-functions-sdk-go/

RUN make build

# Next image - Copy built Go binary into new workspace
# Maybe use FROM scratch for Go binaries??
FROM alpine:3.7

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2019: Intel'

COPY --from=builder /github.com/edgexfoundry/app-functions-sdk-go/examples/simple-filter-xml/simple-filter-xml /usr/local/bin/simple-filter-xml

CMD [ "/usr/local/bin/simple-filter-xml" ]