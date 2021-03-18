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
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"

	mocks2 "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/interfaces/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	msgStr  = "test message"
	path    = "/some-path/foo"
	badPath = "/some-path/bad"
)

func TestHTTPPostPut(t *testing.T) {
	var methodUsed string

	handler := func(w http.ResponseWriter, r *http.Request) {
		methodUsed = r.Method

		if r.URL.EscapedPath() == badPath {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)

		readMsg, _ := ioutil.ReadAll(r.Body)
		_ = r.Body.Close()
		if strings.Compare((string)(readMsg), msgStr) != 0 {
			t.Errorf("Invalid msg received %v, expected %v", readMsg, msgStr)
		}

		if r.Header.Get("Content-type") != "application/json" {
			t.Errorf("Unexpected content-type received %s, expected %s", r.Header.Get("Content-type"), "application/json")
		}
		if r.URL.EscapedPath() != path {
			t.Errorf("Invalid path received %s, expected %s",
				r.URL.EscapedPath(), path)
		}
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	targetUrl, err := url.Parse(ts.URL)
	require.NoError(t, err)

	tests := []struct {
		Name           string
		Path           string
		PersistOnFail  bool
		RetryDataSet   bool
		ExpectedMethod string
	}{
		{"Successful POST", path, true, false, http.MethodPost},
		{"Failed POST no persist", badPath, false, false, http.MethodPost},
		{"Failed POST with persist", badPath, true, true, http.MethodPost},
		{"Successful PUT", path, false, false, http.MethodPut},
		{"Failed PUT no persist", badPath, false, false, http.MethodPut},
		{"Failed PUT with persist", badPath, true, true, http.MethodPut},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			context.SetRetryData(nil)
			methodUsed = ""
			sender := NewHTTPSender(`http://`+targetUrl.Host+test.Path, "", test.PersistOnFail)

			if test.ExpectedMethod == http.MethodPost {
				sender.HTTPPost(context, msgStr)
			} else {
				sender.HTTPPut(context, msgStr)
			}

			assert.Equal(t, test.RetryDataSet, context.RetryData() != nil)
			assert.Equal(t, test.ExpectedMethod, methodUsed)
		})
	}
}

func TestHTTPPostPutWithSecrets(t *testing.T) {
	var methodUsed string

	expectedValue := "my-API-key"

	mockSP := &mocks2.SecretProvider{}
	mockSP.On("GetSecrets", "/path", "header").Return(map[string]string{"Secret-Header-Name": expectedValue}, nil)
	mockSP.On("GetSecrets", "/path", "bogus").Return(nil, errors.New("FAKE NOT FOUND ERROR"))

	dic.Update(di.ServiceConstructorMap{
		bootstrapContainer.SecretProviderName: func(get di.Get) interface{} {
			return mockSP
		},
	})

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		methodUsed = request.Method

		if request.URL.EscapedPath() == badPath {
			writer.WriteHeader(http.StatusNotFound)
			return
		}

		writer.WriteHeader(http.StatusOK)

		actualValue := request.Header.Get("Secret-Header-Name")

		if actualValue != "" {
			// Only validate is key was found in the header
			require.Equal(t, expectedValue, actualValue)
		}
	}))
	defer ts.Close()

	targetUrl, err := url.Parse(ts.URL)
	require.NoError(t, err)

	tests := []struct {
		Name                 string
		Path                 string
		HeaderName           string
		SecretName           string
		SecretPath           string
		ExpectToContinue     bool
		ExpectedErrorMessage string
		ExpectedMethod       string
	}{
		{"unsuccessful POST w/o secret header name", path, "", "header", "/path", false, "HTTP Header Name required when using secrets", ""},
		{"unsuccessful POST w/o secret path", path, "Secret-Header", "header", "", false, "HTTP Header SecretName was provided but no SecretPath was provided", ""},
		{"unsuccessful POST w/o secret name", path, "Secret-Header", "", "/path", false, "SecretPath was specified but no SecretName was provided", ""},
		{"successful POST with secrets", path, "Secret-Header-Name", "header", "/path", true, "", http.MethodPost},
		{"successful POST without secrets", path, "", "", "", true, "", http.MethodPost},
		{"unsuccessful POST with secrets - retrieval fails", path, "Secret-Header", "bogus", "/path", false, "FAKE NOT FOUND ERROR", ""},
		{"unsuccessful PUT w/o secret header name", path, "", "header", "/path", false, "HTTP Header Name required when using secrets", ""},
		{"unsuccessful PUT w/o secret path name", path, "Secret-Header", "header", "", false, "HTTP Header SecretName was provided but no SecretPath was provided", ""},
		{"successful PUT with secrets", path, "Secret-Header", "header", "/path", true, "", http.MethodPut},
		{"successful PUT without secrets", path, "", "", "", true, "", http.MethodPut},
		{"unsuccessful PUT with secrets - retrieval fails", path, "Secret-Header", "bogus", "/path", false, "FAKE NOT FOUND ERROR", ""},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			methodUsed = ""
			sender := NewHTTPSenderWithSecretHeader(
				`http://`+targetUrl.Host+test.Path,
				"",
				false,
				test.HeaderName,
				test.SecretPath,
				test.SecretName)

			var continuePipeline bool
			var err interface{}

			if test.ExpectedMethod == http.MethodPost {
				continuePipeline, err = sender.HTTPPost(context, msgStr)
			} else {
				continuePipeline, err = sender.HTTPPut(context, msgStr)
			}

			assert.Equal(t, test.ExpectToContinue, continuePipeline)
			if !test.ExpectToContinue {
				require.EqualError(t, err.(error), test.ExpectedErrorMessage)
			}
			assert.Equal(t, test.ExpectedMethod, methodUsed)
		})
	}
}

func TestHTTPPostNoParameterPassed(t *testing.T) {
	sender := NewHTTPSender("", "", false)
	continuePipeline, result := sender.HTTPPost(context, nil)

	assert.False(t, continuePipeline, "Pipeline should stop")
	assert.Error(t, result.(error), "Result should be an error")
	assert.Equal(t, "No Data Received", result.(error).Error())
}

func TestHTTPPutNoParameterPassed(t *testing.T) {
	sender := NewHTTPSender("", "", false)
	continuePipeline, result := sender.HTTPPut(context, nil)

	assert.False(t, continuePipeline, "Pipeline should stop")
	assert.Error(t, result.(error), "Result should be an error")
	assert.Equal(t, "No Data Received", result.(error).Error())
}

func TestHTTPPostInvalidParameter(t *testing.T) {
	sender := NewHTTPSender("", "", false)
	// Channels are not marshalable to JSON and generate an error
	data := make(chan int)
	continuePipeline, result := sender.HTTPPost(context, data)

	assert.False(t, continuePipeline, "Pipeline should stop")
	assert.Error(t, result.(error), "Result should be an error")
	assert.Equal(t, "marshaling input data to JSON failed, "+
		"passed in data must be of type []byte, string, or support marshaling to JSON", result.(error).Error())
}

func TestHTTPPutInvalidParameter(t *testing.T) {
	sender := NewHTTPSender("", "", false)
	// Channels are not marshalable to JSON and generate an error
	data := make(chan int)
	continuePipeline, result := sender.HTTPPut(context, data)

	assert.False(t, continuePipeline, "Pipeline should stop")
	assert.Error(t, result.(error), "Result should be an error")
	assert.Equal(t, "marshaling input data to JSON failed, "+
		"passed in data must be of type []byte, string, or support marshaling to JSON", result.(error).Error())
}
