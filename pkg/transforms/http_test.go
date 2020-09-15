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
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/security"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

const (
	msgStr  = "test message"
	path    = "/somepath/foo"
	badPath = "/somepath/bad"
)

var logClient logger.LoggingClient
var config *common.ConfigurationStruct

func TestMain(m *testing.M) {
	logClient = logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")
	config = &common.ConfigurationStruct{}
	m.Run()
}

func TestHTTPPost(t *testing.T) {

	context.CorrelationID = "123"

	handler := func(w http.ResponseWriter, r *http.Request) {

		if r.URL.EscapedPath() == badPath {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)

		readMsg, _ := ioutil.ReadAll(r.Body)
		r.Body.Close()
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

	url, err := url.Parse(ts.URL)
	require.NoError(t, err)

	tests := []struct {
		Name          string
		Path          string
		PersistOnFail bool
		RetryDataSet  bool
	}{
		{"Successful post", path, true, false},
		{"Failed Post no persist", badPath, false, false},
		{"Failed Post with persist", badPath, true, true},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			sender := NewHTTPSender(`http://`+url.Host+test.Path, "", test.PersistOnFail)

			sender.HTTPPost(context, msgStr)

			assert.Equal(t, test.RetryDataSet, context.RetryData != nil)
		})
	}
}

func TestHTTPPostWithSecrets(t *testing.T) {

	expectedValue := "value"
	handler := func(w http.ResponseWriter, r *http.Request) {

		if r.URL.EscapedPath() == badPath {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)

		actualValue := r.Header.Get("Secret-Header-Name")
		if actualValue != "" {
			// Only validate is key was found in the header
			require.Equal(t, expectedValue, actualValue)
		}
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	url, err := url.Parse(ts.URL)
	require.NoError(t, err)

	tests := []struct {
		Name                 string
		Path                 string
		SecretHeaderName     string
		SecretPath           string
		ExpectToContinue     bool
		ExpectedErrorMessage string
	}{
		{"unsuccessful post w/o secret header name", path, "", "/path", false, "SecretPath was specified but no header name was provided"},
		{"unsuccessful post w/o secret path name", path, "Secret-Header-Name", "", false, "HTTP Header Secret Name was provided but no SecretPath was provided"},
		{"successful post with secrets", path, "Secret-Header-Name", "/path", true, ""},
		{"successful post without secrets", path, "", "", true, ""},
		{"unsuccessful post with secrets - retrieval fails", path, "Secret-Header-Name-2", "/path", false, "FAKE NOT FOUND ERROR"},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			var sender HTTPSender

			sender = NewHTTPSenderWithSecretHeader(`http://`+url.Host+test.Path, "", false, test.SecretHeaderName, test.SecretPath)

			continuePipeline, err := sender.HTTPPost(context, msgStr)
			assert.Equal(t, test.ExpectToContinue, continuePipeline)
			if !test.ExpectToContinue {
				require.EqualError(t, err.(error), test.ExpectedErrorMessage)
			}
		})
	}
}

func TestHTTPPostNoParameterPassed(t *testing.T) {

	sender := NewHTTPSender("", "", false)
	continuePipeline, result := sender.HTTPPost(context)

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

type mockSecretClient struct {
}

// NewMockSecretProvider provides a mocked version of the mockSecretClient to avoiding using vault in our tests
func newMockSecretProvider(loggingClient logger.LoggingClient, configuration *common.ConfigurationStruct) security.SecretProvider {
	mockSP := security.NewSecretProvider(logClient, config)
	mockSP.ExclusiveSecretClient = &mockSecretClient{}
	return mockSP
}

// GetSecrets mock implementation of GetSecrets
func (s *mockSecretClient) GetSecrets(path string, keys ...string) (map[string]string, error) {
	fakeDb := map[string]string{"Secret-Header-Name": "value"}
	if _, ok := fakeDb[keys[0]]; ok {
		//do something here
		return fakeDb, nil
	} else {
		return nil, errors.New("FAKE NOT FOUND ERROR")
	}

}

// StoreSecrets mock implementation of StoreSecrets
func (s *mockSecretClient) StoreSecrets(path string, secrets map[string]string) error {
	return nil
}
