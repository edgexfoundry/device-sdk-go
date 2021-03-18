//
// Copyright (c) 2017 Cavium
// Copyright (c) 2021 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package transforms

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
)

type Compression struct {
	gzipWriter *gzip.Writer
	zlibWriter *zlib.Writer
}

// NewCompression creates, initializes and returns a new instance of Compression
func NewCompression() Compression {
	return Compression{}
}

// CompressWithGZIP compresses data received as either a string,[]byte, or json.Marshaller using gzip algorithm
// and returns a base64 encoded string as a []byte.
func (compression *Compression) CompressWithGZIP(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}
	ctx.LoggingClient().Debug("Compression with GZIP")
	rawData, err := util.CoerceType(data)
	if err != nil {
		return false, err
	}
	var buf bytes.Buffer

	if compression.gzipWriter == nil {
		compression.gzipWriter = gzip.NewWriter(&buf)
	} else {
		compression.gzipWriter.Reset(&buf)
	}

	_, err = compression.gzipWriter.Write(rawData)
	if err != nil {
		return false, fmt.Errorf("unable to write GZIP data: %s", err.Error())
	}

	err = compression.gzipWriter.Close()
	if err != nil {
		return false, fmt.Errorf("unable to close GZIP data: %s", err.Error())
	}

	// Set response "content-type" header to "text/plain"
	ctx.SetResponseContentType(clients.ContentTypeText)

	return true, bytesBufferToBase64(buf)

}

// CompressWithZLIB compresses data received as either a string,[]byte, or json.Marshaller using zlib algorithm
// and returns a base64 encoded string as a []byte.
func (compression *Compression) CompressWithZLIB(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	if data == nil {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}
	ctx.LoggingClient().Debug("Compression with ZLIB")
	byteData, err := util.CoerceType(data)
	if err != nil {
		return false, err
	}
	var buf bytes.Buffer

	if compression.zlibWriter == nil {
		compression.zlibWriter = zlib.NewWriter(&buf)
	} else {
		compression.zlibWriter.Reset(&buf)
	}

	_, err = compression.zlibWriter.Write(byteData)
	if err != nil {
		return false, fmt.Errorf("unable to write ZLIB data: %s", err.Error())
	}

	err = compression.zlibWriter.Close()
	if err != nil {
		return false, fmt.Errorf("unable to close ZLIB data: %s", err.Error())
	}

	// Set response "content-type" header to "text/plain"
	ctx.SetResponseContentType(clients.ContentTypeText)

	return true, bytesBufferToBase64(buf)

}

func bytesBufferToBase64(buf bytes.Buffer) []byte {
	dst := make([]byte, base64.StdEncoding.EncodedLen(buf.Len()))
	base64.StdEncoding.Encode(dst, buf.Bytes())
	return dst
}
