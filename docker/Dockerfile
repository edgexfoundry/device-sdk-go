#
# Copyright (c) 2018
# Intel
#
# SPDX-License-Identifier: Apache-2.0
#
FROM golang:1.11-alpine AS builder

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2018: Intel'

RUN sed -e 's/dl-cdn[.]alpinelinux.org/nl.alpinelinux.org/g' -i~ /etc/apk/repositories

# add git for go modules
RUN apk add --update --no-cache make git gcc libc-dev

WORKDIR /github.com/edgexfoundry/device-sdk-go

COPY . /github.com/edgexfoundry/device-sdk-go/

RUN make build

# Next image - Copy built Go binary into new workspace
# Maybe use FROM scratch for Go binaries??
FROM alpine:3.7

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2018: Intel'

COPY --from=builder /github.com/edgexfoundry/device-sdk-go/example/cmd/device-simple/device-simple /usr/local/bin/device-simple

CMD [ "/usr/local/bin/device-simple" ]
