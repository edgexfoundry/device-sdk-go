#
# Copyright (c) 2018, 2019 Intel
#
# SPDX-License-Identifier: Apache-2.0
#
FROM golang:1.11-alpine AS builder

ENV GO111MODULE=on
WORKDIR /go/src/github.com/edgexfoundry/device-sdk-go

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2018, 2019: Intel'

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

# add git for go modules
RUN apk update && apk add make git

COPY go.mod .

RUN go mod download

COPY . .

RUN make build

# Next image - Copy built Go binary into new workspace
FROM scratch

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2018, 2019: Intel'

ENV APP_PORT=49990
#expose command data port
EXPOSE $APP_PORT

WORKDIR /
COPY --from=builder /go/src/github.com/edgexfoundry/device-sdk-go/example/cmd/device-simple/device-simple /usr/local/bin/device-simple
COPY --from=builder /go/src/github.com/edgexfoundry/device-sdk-go/example/cmd/device-simple/res/docker/configuration.toml /res/docker/configuration.toml

CMD [ "/usr/local/bin/device-simple","--registry","--profile=docker","--confdir=/res"]
