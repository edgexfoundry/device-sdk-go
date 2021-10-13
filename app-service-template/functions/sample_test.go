// TODO: Change Copyright to your company if open sourcing or remove header
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

package functions

import (
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/pkg/interfaces"

	"github.com/edgexfoundry/go-mod-core-contracts/v2/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/common"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file contains example of how to unit test pipeline functions
// TODO: Change these sample unit tests to test your custom type and function(s)

var appContext interfaces.AppFunctionContext

func TestMain(m *testing.M) {
	//
	// This can be changed to a real logger when needing more debug information output to the console
	// lc := logger.NewClient("testing", "DEBUG")
	//
	lc := logger.NewMockClient()
	correlationId := uuid.New().String()

	// NewAppFuncContextForTest creates a context with basic dependencies for unit testing with the passed in logger
	// If more additional dependencies (such as mock clients) are required, then use
	// NewAppFuncContext(correlationID string, dic *di.Container) and pass in an initialized DIC (dependency injection container)
	appContext = pkg.NewAppFuncContextForTest(correlationId, lc)
}

func TestSample_LogEventDetails(t *testing.T) {
	expectedEvent := createTestEvent(t)
	expectedContinuePipeline := true

	target := NewSample()
	actualContinuePipeline, actualEvent := target.LogEventDetails(appContext, expectedEvent)

	assert.Equal(t, expectedContinuePipeline, actualContinuePipeline)
	assert.Equal(t, expectedEvent, actualEvent)
}

func TestSample_ConvertEventToXML(t *testing.T) {
	event := createTestEvent(t)
	expectedXml, _ := event.ToXML()
	expectedContinuePipeline := true

	target := NewSample()
	actualContinuePipeline, actualXml := target.ConvertEventToXML(appContext, event)

	assert.Equal(t, expectedContinuePipeline, actualContinuePipeline)
	assert.Equal(t, expectedXml, actualXml)

}

func TestSample_OutputXML(t *testing.T) {
	testEvent := createTestEvent(t)
	xml, _ := testEvent.ToXML()
	expectedContinuePipeline := false
	expectedContentType := common.ContentTypeXML

	target := NewSample()
	actualContinuePipeline, result := target.OutputXML(appContext, xml)
	actualContentType := appContext.ResponseContentType()

	assert.Equal(t, expectedContinuePipeline, actualContinuePipeline)
	assert.Nil(t, result)
	assert.Equal(t, expectedContentType, actualContentType)
}

func createTestEvent(t *testing.T) dtos.Event {
	profileName := "MyProfile"
	deviceName := "MyDevice"
	sourceName := "MySource"
	resourceName := "MyResource"

	event := dtos.NewEvent(profileName, deviceName, sourceName)
	err := event.AddSimpleReading(resourceName, common.ValueTypeInt32, int32(1234))
	require.NoError(t, err)

	event.Tags = map[string]interface{}{
		"WhereAmI": "NotKansas",
	}

	return event
}
