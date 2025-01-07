//
// Copyright (c) 2019 Intel Corporation
// Copyright (C) 2020-2025 IOTech Ltd
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

package http

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/common"
	"github.com/stretchr/testify/assert"

	"github.com/edgexfoundry/device-sdk-go/v4/internal/config"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/container"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

var (
	handlerFunc = func(c echo.Context) error {
		return c.String(http.StatusOK, "OK")
	}
)

func TestAddRoute(t *testing.T) {

	tests := []struct {
		Name          string
		Route         string
		Method        string
		ErrorExpected bool
	}{
		{"Success", "/api/v2/test", http.MethodPost, false},
		{"Reserved Route", common.ApiDiscoveryRoute, "", true},
	}

	lc := logger.NewMockClient()
	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{}
		},
	})

	for _, test := range tests {
		r := echo.New()
		controller := NewRestController(r, dic, uuid.NewString())
		controller.InitRestRoutes(dic)

		err := controller.AddRoute(test.Route, handlerFunc, []string{test.Method})
		if test.ErrorExpected {
			assert.Error(t, err, "Expected an error")
		} else {
			if !assert.NoError(t, err, "Unexpected an error") {
				t.Fatal()
			}

			req := httptest.NewRequest(test.Method, test.Route, nil)
			rec := httptest.NewRecorder()
			c := r.NewContext(req, rec)
			// Find the matched handler function from router with the matching method and url path
			r.Router().Find(test.Method, test.Route, c)
			// Apply the handler function to echo.Context
			handlerErr := c.Handler()(c)
			assert.NoError(t, handlerErr)

			// Have to skip all the reserved routes that have previously been added.
			if controller.reservedRoutes[test.Route] {
				return
			}

			assert.Equal(t, test.Route, c.Path())
			assert.Equal(t, http.StatusOK, c.Response().Status)

			if body, err := io.ReadAll(rec.Body); err == nil {
				assert.Equal(t, "OK", string(body), "unexpected handler function response")
			} else {
				assert.NoError(t, err)
			}
		}
	}
}

func TestInitRestRoutes(t *testing.T) {
	lc := logger.NewMockClient()
	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &config.ConfigurationStruct{}
		},
	})
	r := echo.New()
	controller := NewRestController(r, dic, uuid.NewString())
	controller.InitRestRoutes(dic)

	// Traverse all registered routes for the router
	for _, route := range r.Routes() {
		path := route.Path

		// Verify the route is reserved by attempting to add it as 'external' route.
		// If tests fails then the route was not added to the reserved list
		err := controller.AddRoute(path, func(c echo.Context) error { return nil }, nil)
		assert.Error(t, err, path, fmt.Sprintf("Expected error for '%s'", path))
	}
}
