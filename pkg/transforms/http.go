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

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/appcontext"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
)

// HTTPSender ...
type HTTPSender struct {
	URL              string
	MimeType         string
	PersistOnError   bool
	SecretHeaderName string
	SecretPath       string
}

// NewHTTPSender creates, initializes and returns a new instance of HTTPSender
func NewHTTPSender(url string, mimeType string, persistOnError bool) HTTPSender {
	return HTTPSender{
		URL:            url,
		MimeType:       mimeType,
		PersistOnError: persistOnError,
	}
}

// NewHTTPSenderWithSecretHeader creates, initializes and returns a new instance of HTTPSender configured to use a secret header
func NewHTTPSenderWithSecretHeader(url string, mimeType string, persistOnError bool, httpHeaderSecretName string, secretPath string) HTTPSender {
	return HTTPSender{
		URL:              url,
		MimeType:         mimeType,
		PersistOnError:   persistOnError,
		SecretHeaderName: httpHeaderSecretName,
		SecretPath:       secretPath,
	}
}

// HTTPPost will send data from the previous function to the specified Endpoint via http POST.
// If no previous function exists, then the event that triggered the pipeline will be used.
// An empty string for the mimetype will default to application/json.
func (sender HTTPSender) HTTPPost(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	return sender.httpSend(edgexcontext, params, http.MethodPost)
}

// HTTPPut will send data from the previous function to the specified Endpoint via http PUT.
// If no previous function exists, then the event that triggered the pipeline will be used.
// An empty string for the mimetype will default to application/json.
func (sender HTTPSender) HTTPPut(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	return sender.httpSend(edgexcontext, params, http.MethodPut)
}

func (sender HTTPSender) httpSend(edgexcontext *appcontext.Context, params []interface{}, method string) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}

	if sender.MimeType == "" {
		sender.MimeType = "application/json"
	}

	exportData, err := util.CoerceType(params[0])
	if err != nil {
		return false, err
	}

	usingSecrets, err := sender.determineIfUsingSecrets()
	if err != nil {
		return false, err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, sender.URL, bytes.NewReader(exportData))
	if err != nil {
		return false, err
	}
	var theSecrets map[string]string
	if usingSecrets {
		theSecrets, err = edgexcontext.GetSecrets(sender.SecretPath, sender.SecretHeaderName)
		if err != nil {
			return false, err
		}
		req.Header.Set(sender.SecretHeaderName, theSecrets[sender.SecretHeaderName])
	}

	req.Header.Set("Content-Type", sender.MimeType)

	edgexcontext.LoggingClient.Debug("POSTing data")
	response, err := client.Do(req)
	if err != nil {
		sender.setRetryData(edgexcontext, exportData)
		return false, err
	}
	defer response.Body.Close()
	edgexcontext.LoggingClient.Debug(fmt.Sprintf("Response: %s", response.Status))
	edgexcontext.LoggingClient.Debug(fmt.Sprintf("Sent data: %s", string(exportData)))
	bodyBytes, errReadingBody := ioutil.ReadAll(response.Body)
	if errReadingBody != nil {
		sender.setRetryData(edgexcontext, exportData)
		return false, errReadingBody
	}

	edgexcontext.LoggingClient.Trace("Data exported", "Transport", "HTTP", clients.CorrelationHeader, edgexcontext.CorrelationID)

	// continues the pipeline if we get a 2xx response, stops pipeline if non-2xx response
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		sender.setRetryData(edgexcontext, exportData)
		return false, fmt.Errorf("export failed with %d HTTP status code", response.StatusCode)
	}

	return true, bodyBytes
}

func (sender HTTPSender) determineIfUsingSecrets() (bool, error) {
	//check if one field but not others are provided for secrets
	if sender.SecretPath != "" && sender.SecretHeaderName == "" {
		return false, errors.New("SecretPath was specified but no header name was provided")
	}
	if sender.SecretHeaderName != "" && sender.SecretPath == "" {
		return false, errors.New("HTTP Header Secret Name was provided but no SecretPath was provided")
	}

	// not using secrets if both are blank
	if sender.SecretHeaderName == "" && sender.SecretPath == "" {
		return false, nil
	}
	// using secrets, all required fields are provided
	return true, nil

}

func (sender HTTPSender) setRetryData(ctx *appcontext.Context, exportData []byte) {
	if sender.PersistOnError {
		ctx.RetryData = exportData
	}
}
