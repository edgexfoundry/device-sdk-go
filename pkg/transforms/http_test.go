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

	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logging"

	"github.com/edgexfoundry/app-functions-sdk-go/pkg/excontext"
)

var lc logger.LoggingClient

func init() {
	lc = logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")
}
func TestHTTPPost(t *testing.T) {
	const (
		msgStr = "test message"
		path   = "/somepath/foo"
	)

	handler := func(w http.ResponseWriter, r *http.Request) {
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
	if err != nil {
		t.Fatal("Could not parse url")
	}

	ctx := excontext.Context{
		LoggingClient: lc,
	}
	sender := HTTPSender{
		URL: `http://` + url.Host + path,
	}
	sender.HTTPPost(ctx, msgStr)
}

func TestHTTPPostNoParameterPassed(t *testing.T) {
	ctx := excontext.Context{
		LoggingClient: lc,
	}
	sender := HTTPSender{}
	continuePipeline, result := sender.HTTPPost(ctx)
	if continuePipeline != false {
		t.Fatal("Pipeline should stop")
	}
	if result.(error).Error() != "No Data Received" {
		t.Fatal("Should have an error when no parameter was passed")
	}
}
func TestHTTPPostInvalidParameter(t *testing.T) {
	ctx := excontext.Context{
		LoggingClient: lc,
	}
	sender := HTTPSender{}
	data := "HELLO"
	continuePipeline, result := sender.HTTPPost(ctx, ([]byte)(data))
	if continuePipeline != false {
		t.Fatal("Pipeline should stop")
	}
	if result.(error).Error() != "Unexpected type received" {
		t.Fatal("Should have an error when no parameter was passed")
	}
}
func TestHTTPPostNonExistentEndpoint(t *testing.T) {
	ctx := excontext.Context{
		LoggingClient: lc,
	}
	sender := HTTPSender{
		URL: "http://idontexist/",
	}
	continuePipeline, result := sender.HTTPPost(ctx, "data")
	if continuePipeline != false {
		t.Fatal("Pipeline should stop")
	}
	if !strings.Contains(result.(error).Error(), "no such host") {
		t.Fatal("Should have an error from http post that does not find host")
	}
}
