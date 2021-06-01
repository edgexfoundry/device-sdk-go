//
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
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/util"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients"
)

// HTTPSender ...
type HTTPSender struct {
	URL                 string
	MimeType            string
	PersistOnError      bool
	ContinueOnSendError bool
	ReturnInputData     bool
	HttpHeaderName      string
	SecretName          string
	SecretPath          string
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
func NewHTTPSenderWithSecretHeader(url string, mimeType string, persistOnError bool, headerName string, secretPath string, secretName string) HTTPSender {
	return HTTPSender{
		URL:            url,
		MimeType:       mimeType,
		PersistOnError: persistOnError,
		HttpHeaderName: headerName,
		SecretPath:     secretPath,
		SecretName:     secretName,
	}
}

// HTTPPost will send data from the previous function to the specified Endpoint via http POST.
// If no previous function exists, then the event that triggered the pipeline will be used.
// An empty string for the mimetype will default to application/json.
func (sender HTTPSender) HTTPPost(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	return sender.httpSend(ctx, data, http.MethodPost)
}

// HTTPPut will send data from the previous function to the specified Endpoint via http PUT.
// If no previous function exists, then the event that triggered the pipeline will be used.
// An empty string for the mimetype will default to application/json.
func (sender HTTPSender) HTTPPut(ctx interfaces.AppFunctionContext, data interface{}) (bool, interface{}) {
	return sender.httpSend(ctx, data, http.MethodPut)
}

func (sender HTTPSender) httpSend(ctx interfaces.AppFunctionContext, data interface{}, method string) (bool, interface{}) {
	lc := ctx.LoggingClient()

	lc.Debug("HTTP Exporting")

	if data == nil {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}

	if sender.PersistOnError && sender.ContinueOnSendError {
		return false, errors.New("PersistOnError & ContinueOnSendError can not both be set to true for HTTP Export")
	}

	if sender.ContinueOnSendError && !sender.ReturnInputData {
		return false, errors.New("ContinueOnSendError can only be used in conjunction ReturnInputData for multiple HTTP Export")
	}

	if sender.MimeType == "" {
		sender.MimeType = "application/json"
	}

	exportData, err := util.CoerceType(data)
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
		theSecrets, err = ctx.GetSecret(sender.SecretPath, sender.SecretName)
		if err != nil {
			return false, err
		}

		lc.Debugf("Setting HTTP Header '%s' with secret value from SecretStore at path='%s' & name='%s",
			sender.HttpHeaderName,
			sender.SecretPath,
			sender.SecretName)

		req.Header.Set(sender.HttpHeaderName, theSecrets[sender.SecretName])
	}

	req.Header.Set("Content-Type", sender.MimeType)

	ctx.LoggingClient().Debugf("POSTing data to %s", sender.URL)

	response, err := client.Do(req)
	// Pipeline continues if we get a 2xx response, non-2xx response may stop pipeline
	if err != nil || response.StatusCode < 200 || response.StatusCode >= 300 {
		if err == nil {
			err = fmt.Errorf("export failed with %d HTTP status code", response.StatusCode)
		} else {
			err = fmt.Errorf("export failed: %w", err)
		}

		// If continuing on send error then can't be persisting on error since Store and Forward retries starting
		// with the function that failed and stopped the execution of the pipeline.
		if !sender.ContinueOnSendError {
			sender.setRetryData(ctx, exportData)
			return false, err
		}

		// Continuing pipeline on error
		// This is in support of sending to multiple export destinations by chaining export functions in the pipeline.
		ctx.LoggingClient().Errorf("Continuing pipeline on error: %s", err.Error())

		// Return the input data since must have some data for the next function to operate on.
		return true, data
	}

	ctx.LoggingClient().Debugf("Sent %s bytes of data. Response status is %s", len(exportData), response.Status)
	ctx.LoggingClient().Trace("Data exported", "Transport", "HTTP", clients.CorrelationHeader, ctx.CorrelationID)

	// This allows multiple HTTP Exports to be chained in the pipeline to send the same data to different destinations
	// Don't need to read the response data since not going to return it so just return now.
	if sender.ReturnInputData {
		return true, data
	}

	defer func() { _ = response.Body.Close() }()
	responseData, errReadingBody := ioutil.ReadAll(response.Body)
	if errReadingBody != nil {
		// Can't have ContinueOnSendError=true when ReturnInputData=false, so no need to check for it here
		sender.setRetryData(ctx, exportData)
		return false, errReadingBody
	}

	return true, responseData
}

func (sender HTTPSender) determineIfUsingSecrets() (bool, error) {
	// not using secrets if both are empty
	if len(sender.SecretPath) == 0 && len(sender.SecretName) == 0 {
		if len(sender.HttpHeaderName) == 0 {
			return false, nil
		}

		return false, errors.New("SecretPath & SecretName must be specified when HTTP Header Name is specified")
	}

	//check if one field but not others are provided for secrets
	if len(sender.SecretPath) != 0 && len(sender.SecretName) == 0 {
		return false, errors.New("SecretPath was specified but no SecretName was provided")
	}
	if len(sender.SecretName) != 0 && len(sender.SecretPath) == 0 {
		return false, errors.New("HTTP Header SecretName was provided but no SecretPath was provided")
	}

	if len(sender.HttpHeaderName) == 0 {
		return false, errors.New("HTTP Header Name required when using secrets")
	}

	// using secrets, all required fields are provided
	return true, nil
}

func (sender HTTPSender) setRetryData(ctx interfaces.AppFunctionContext, exportData []byte) {
	if sender.PersistOnError {
		ctx.SetRetryData(exportData)
	}
}
