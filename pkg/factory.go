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

package pkg

import (
	"fmt"

	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/app"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/appfunction"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/bootstrap/container"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/common"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"
)

// NewAppService creates and returns a new ApplicationService with the default TargetType
func NewAppService(serviceKey string) (interfaces.ApplicationService, bool) {
	return NewAppServiceWithTargetType(serviceKey, nil)
}

// NewAppServiceWithTargetType creates and returns a new ApplicationService with the specified TargetType
func NewAppServiceWithTargetType(serviceKey string, targetType interface{}) (interfaces.ApplicationService, bool) {
	service := app.NewService(serviceKey, targetType, interfaces.ProfileSuffixPlaceholder)
	if err := service.Initialize(); err != nil {
		err = fmt.Errorf("initialization failed: %s", err.Error())
		service.LoggingClient().Errorf("App Service %s", err.Error())
		return nil, false
	}

	return service, true
}

// NewAppFuncContextForTest creates and returns a new AppFunctionContext to be used in unit tests for custom pipeline functions
func NewAppFuncContextForTest(correlationID string, lc logger.LoggingClient) interfaces.AppFunctionContext {
	dic := di.NewContainer(di.ServiceConstructorMap{
		bootstrapContainer.LoggingClientInterfaceName: func(get di.Get) interface{} {
			return lc
		},
		container.ConfigurationName: func(get di.Get) interface{} {
			return &common.ConfigurationStruct{}
		},
	})
	return appfunction.NewContext(correlationID, dic, "")
}
