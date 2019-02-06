#
# Copyright (c) 2018
# Intel
#
# SPDX-License-Identifier: Apache-2.0
#
FROM golang:1.11.2-alpine3.7 AS builder

LABEL license='SPDX-License-Identifier: Apache-2.0' \
  copyright='Copyright (c) 2018: Intel'

# The main mirrors are giving us timeout issues on builds periodically.
# So we can try these.
RUN echo http://nl.alpinelinux.org/alpine/v3.7/main > /etc/apk/repositories; \
  echo http://nl.alpinelinux.org/alpine/v3.7/community >> /etc/apk/repositories

# add git for go modules
RUN apk update && apk add --no-cache make git gcc libc-dev

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
