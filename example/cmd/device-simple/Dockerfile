#
# Copyright (c) 2020 Intel
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

ARG BASE=golang:1.15-alpine3.12
FROM ${BASE} AS builder

ARG MAKE='make build'

WORKDIR /device-sdk-go

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2020: Intel'

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

# add git for go modules
RUN apk add --update --no-cache make git

COPY . .

RUN ${MAKE}

# Next image - Copy built Go binary into new workspace
FROM scratch

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2020: Intel'

ENV APP_PORT=49990
#expose command data port
EXPOSE $APP_PORT

WORKDIR /
COPY --from=builder /device-sdk-go/example/cmd/device-simple/device-simple /usr/local/bin/device-simple
COPY --from=builder /device-sdk-go/example/cmd/device-simple/res/configuration.toml /res/configuration.toml
COPY --from=builder /device-sdk-go/example/cmd/device-simple/res/Simple-Driver.yaml /res/Simple-Driver.yaml
COPY --from=builder /device-sdk-go/example/cmd/device-simple/res/on.png /res/on.png
COPY --from=builder /device-sdk-go/example/cmd/device-simple/res/off.jpg /res/off.jpg

ENTRYPOINT ["/usr/local/bin/device-simple"]
CMD ["-cp=consul.http://edgex-core-consul:8500", "--registry", "--confdir=/res"]
