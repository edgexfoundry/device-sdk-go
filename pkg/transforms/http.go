//
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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
)

// HTTPSender ...
type HTTPSender struct {
	URL      string
	MimeType string
}

// HTTPPost ...
func (sender HTTPSender) HTTPPost(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}
	if sender.MimeType == "" {
		sender.MimeType = "application/json"
	}
	if result, ok := params[0].(string); ok {
		edgexcontext.LoggingClient.Info("POSTing data")
		response, err := http.Post(sender.URL, sender.MimeType, bytes.NewReader(([]byte)(result)))
		if err != nil {
			//LoggingClient.Error(err.Error())
			return false, err
		}
		defer response.Body.Close()
		edgexcontext.LoggingClient.Info(fmt.Sprintf("Response: %s", response.Status))
		edgexcontext.LoggingClient.Debug(fmt.Sprintf("Sent data: %s", result))
		bodyBytes, errReadingBody := ioutil.ReadAll(response.Body)
		if errReadingBody != nil {
			return false, errReadingBody
		}
		// continues the pipeline if we get a 2xx response, stops pipeline if non-2xx response
		return response.StatusCode >= 200 && response.StatusCode < 300, bodyBytes
	}

	return false, errors.New("Unexpected type received")
}
