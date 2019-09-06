// +build mongoRunning

/*******************************************************************************
 * Copyright 2019 Dell Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software distributed under the License
 * is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express
 * or implied. See the License for the specific language governing permissions and limitations under
 * the License.
 *******************************************************************************/

// This test will only be executed if the tag mongoRunning is added when running
// the tests with a command like:
// go test -tags mongoRunning

package mongo

import (
	"reflect"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var TestHost = "localhost"
var TestPort = 27017
var TestTimeout = 5000
var TestDatabaseName = "test"
var TestBatchSize = 1337

var TestValidNoAuthConfig = db.Configuration{
	Type:         db.MongoDB,
	Host:         TestHost,
	Port:         TestPort,
	Timeout:      TestTimeout,
	DatabaseName: TestDatabaseName,
	BatchSize:    TestBatchSize,
}

var TestTimeoutConfig = db.Configuration{
	Type:         db.MongoDB,
	Host:         TestHost,
	Port:         TestPort,
	Timeout:      1,
	DatabaseName: TestDatabaseName,
	BatchSize:    TestBatchSize,
}

var TestID = primitive.NewObjectID().Hex()
var TestAppServiceKey = uuid.New().String()
var TestPayload = []byte("brandon was here")
var TestPipelinePosition = 0
var TestVersion = "1"

var TestStoredObject = models.NewStoredObject(TestID, TestAppServiceKey, TestPayload, TestPipelinePosition, TestVersion)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name          string
		config        db.Configuration
		expectedError bool
	}{
		{"Success, no auth", TestValidNoAuthConfig, false},
		{"Timed out", TestTimeoutConfig, true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewClient(test.config)

			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}
		})
	}
}

func TestClient_CRUD(t *testing.T) {
	tests := []struct {
		name             string
		config           db.Configuration
		expectedError    bool
		expectedErrorVal error
	}{
		{"Success, no auth", TestValidNoAuthConfig, false, nil},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client, _ := NewClient(test.config)

			err := client.Store(TestStoredObject)

			if test.expectedError && err == nil {
				t.Error("Expected an error")
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if test.expectedErrorVal != nil && err != nil {
				if test.expectedErrorVal.Error() != err.Error() {
					t.Errorf("Observed error doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedErrorVal.Error(), err.Error())
					return
				}
			}

			retrievedValSlice, err := client.RetrieveFromStore(TestAppServiceKey)

			if retrievedValSlice == nil {
				t.Error("No objects retrieved from store")
				return
			}

			if !reflect.DeepEqual(retrievedValSlice[0], TestStoredObject) {
				t.Errorf("Return value doesn't match expected.\nExpected: %v\nActual: %v\n", TestStoredObject, retrievedValSlice[0])
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			if test.expectedErrorVal != nil && err != nil {
				if test.expectedErrorVal.Error() != err.Error() {
					t.Errorf("Observed error doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedErrorVal.Error(), err.Error())
					return
				}
			}

			updatedTestObject := TestStoredObject
			updatedTestObject.RetryCount = 10

			err = client.UpdateRetryCount(updatedTestObject.ID, updatedTestObject.RetryCount)

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			retrievedValSlice, err = client.RetrieveFromStore(TestAppServiceKey)

			if retrievedValSlice == nil {
				t.Error("No objects retrieved from store")
				return
			}

			if !reflect.DeepEqual(retrievedValSlice[0], updatedTestObject) {
				t.Errorf("Return value doesn't match expected.\nExpected: %v\nActual: %v\n", updatedTestObject, retrievedValSlice[0])
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			updatedTestObject.Version = "2"

			err = client.Update(updatedTestObject)

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			retrievedValSlice, err = client.RetrieveFromStore(updatedTestObject.AppServiceKey)

			if retrievedValSlice == nil {
				t.Error("No objects retrieved from store")
				return
			}

			if !reflect.DeepEqual(retrievedValSlice[0], updatedTestObject) {
				t.Errorf("Return value doesn't match expected.\nExpected: %v\nActual: %v\n", updatedTestObject, retrievedValSlice[0])
				return
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			err = client.RemoveFromStore(TestID)

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}

			shouldBeEmpty, err := client.RetrieveFromStore(TestAppServiceKey)

			if len(shouldBeEmpty) != 0 {
				t.Error("Objects should be deleted")
			}

			if !test.expectedError && err != nil {
				t.Errorf("Unexpectedly encountered error: %s", err.Error())
				return
			}
		})
	}
}
