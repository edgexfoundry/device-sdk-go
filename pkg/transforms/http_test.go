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
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/stretchr/testify/assert"
)

func TestHTTPPost(t *testing.T) {
	const (
		msgStr  = "test message"
		path    = "/somepath/foo"
		badPath = "/somepath/bad"
	)

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
