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
	"encoding/json"
	"errors"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
)

type Compression struct {
	gzipWriter *gzip.Writer
	zlibWriter *zlib.Writer
}

func (compression *Compression) GZIPTransform(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}
	edgexcontext.LoggingClient.Debug("Compression with GZIP")
	var data []byte
	var err error
	switch params[0].(type) {
	case string:
		input := params[0].(string)
		data = []byte(input)

	case []byte:
		data = params[0].([]byte)

	case json.Marshaler:
		marshaler := params[0].(json.Marshaler)
		data, err = marshaler.MarshalJSON()
		if err != nil {
			return false, errors.New("Marshaling input data to JSON failed")
		}
	default:
		return false, errors.New("Unexpected type received - passed in data must be of type []byte, string or implement json.Marshaler")
	}
	var buf bytes.Buffer

	if compression.gzipWriter == nil {
		compression.gzipWriter = gzip.NewWriter(&buf)
	} else {
		compression.gzipWriter.Reset(&buf)
	}

	compression.gzipWriter.Write([]byte(data))
	compression.gzipWriter.Close()

	return true, bytesBufferToBase64(buf)

}

func (compression *Compression) ZLIBTransform(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}
	edgexcontext.LoggingClient.Debug("Compression with ZLIB")
	var data []byte
	var err error
	switch params[0].(type) {
	case string:
		input := params[0].(string)
		data = []byte(input)

	case []byte:
		data = params[0].([]byte)

	case json.Marshaler:
		marshaler := params[0].(json.Marshaler)
		data, err = marshaler.MarshalJSON()
		if err != nil {
			return false, errors.New("Marshaling input data to JSON failed")
		}
	default:
		return false, errors.New("Unexpected type received - passed in data must be of type []byte, string or implement json.Marshaler")
	}
	var buf bytes.Buffer

	if compression.zlibWriter == nil {
		compression.zlibWriter = zlib.NewWriter(&buf)
	} else {
		compression.zlibWriter.Reset(&buf)
	}

	compression.zlibWriter.Write([]byte(data))
	compression.zlibWriter.Close()

	return true, bytesBufferToBase64(buf)

}

func bytesBufferToBase64(buf bytes.Buffer) []byte {
	dst := make([]byte, base64.StdEncoding.EncodedLen(buf.Len()))
	base64.StdEncoding.Encode(dst, buf.Bytes())
	return dst
}
