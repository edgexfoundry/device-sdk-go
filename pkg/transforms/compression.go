//
// Copyright (c) 2017 Cavium
// Copyright (c) 2019 Intel Corporation
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

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/util"
	"github.com/edgexfoundry/go-mod-core-contracts/clients"
)

type Compression struct {
	gzipWriter *gzip.Writer
	zlibWriter *zlib.Writer
}

// NewCompression creates, initializes and returns a new instance of Compression
func NewCompression() Compression {
	return Compression{}
}

// CompressWithGZIP compresses data received as either a string,[]byte, or json.Marshaler using gzip algorithm
// and returns a base64 encoded string as a []byte.
func (compression *Compression) CompressWithGZIP(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}
	edgexcontext.LoggingClient.Debug("Compression with GZIP")
	data, err := util.CoerceType(params[0])
	if err != nil {
		return false, err
	}
	var buf bytes.Buffer

	if compression.gzipWriter == nil {
		compression.gzipWriter = gzip.NewWriter(&buf)
	} else {
		compression.gzipWriter.Reset(&buf)
	}

	compression.gzipWriter.Write([]byte(data))
	compression.gzipWriter.Close()

	// Set response "content-type" header to "text/plain"
	edgexcontext.ResponseContentType = clients.ContentTypeText

	return true, bytesBufferToBase64(buf)

}

// CompressWithZLIB compresses data received as either a string,[]byte, or json.Marshaler using zlib algorithm
// and returns a base64 encoded string as a []byte.
func (compression *Compression) CompressWithZLIB(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}
	edgexcontext.LoggingClient.Debug("Compression with ZLIB")
	data, err := util.CoerceType(params[0])
	if err != nil {
		return false, err
	}
	var buf bytes.Buffer

	if compression.zlibWriter == nil {
		compression.zlibWriter = zlib.NewWriter(&buf)
	} else {
		compression.zlibWriter.Reset(&buf)
	}

	compression.zlibWriter.Write([]byte(data))
	compression.zlibWriter.Close()

	// Set response "content-type" header to "text/plain"
	edgexcontext.ResponseContentType = clients.ContentTypeText

	return true, bytesBufferToBase64(buf)

}

func bytesBufferToBase64(buf bytes.Buffer) []byte {
	dst := make([]byte, base64.StdEncoding.EncodedLen(buf.Len()))
	base64.StdEncoding.Encode(dst, buf.Bytes())
	return dst
}
