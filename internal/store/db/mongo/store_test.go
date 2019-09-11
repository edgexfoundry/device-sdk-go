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

	"github.com/google/uuid"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/contracts"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db"
)

const (
	TestHost         = "localhost"
	TestPort         = 27017
	TestTimeout      = 5000
	TestDatabaseName = "test"
	TestBatchSize    = 1337

	TestPipelinePosition = 0
	TestVersion          = "1"
)

var TestAppServiceKey = uuid.New().String()
var TestPayload = []byte("brandon was here")

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

var TestStoredObject = contracts.NewStoredObject(TestAppServiceKey, TestPayload, TestPipelinePosition, TestVersion)

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
				t.Fatal("Expected an error")
			}

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
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

			id, err := client.Store(TestStoredObject)
			TestStoredObject.ID = id

			if test.expectedError && err == nil {
				t.Fatal("Expected an error")
			}

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}

			if test.expectedErrorVal != nil && err != nil {
				if test.expectedErrorVal.Error() != err.Error() {
					t.Fatalf("Observed error doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedErrorVal.Error(), err.Error())
				}
			}

			retrievedValSlice, err := client.RetrieveFromStore(TestAppServiceKey)

			if retrievedValSlice == nil {
				t.Fatal("No objects retrieved from store")
			}

			if !reflect.DeepEqual(retrievedValSlice[0], TestStoredObject) {
				t.Fatalf("Return value doesn't match expected.\nExpected: %v\nActual: %v\n", TestStoredObject, retrievedValSlice[0])
			}

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}

			if test.expectedErrorVal != nil && err != nil {
				if test.expectedErrorVal.Error() != err.Error() {
					t.Fatalf("Observed error doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedErrorVal.Error(), err.Error())
				}
			}

			updatedTestObject := TestStoredObject
			updatedTestObject.RetryCount = 10

			err = client.UpdateRetryCount(updatedTestObject.ID, updatedTestObject.RetryCount)

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}

			retrievedValSlice, err = client.RetrieveFromStore(TestAppServiceKey)

			if retrievedValSlice == nil {
				t.Fatal("No objects retrieved from store")
			}

			if !reflect.DeepEqual(retrievedValSlice[0], updatedTestObject) {
				t.Fatalf("Return value doesn't match expected.\nExpected: %v\nActual: %v\n", updatedTestObject, retrievedValSlice[0])
			}

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}

			updatedTestObject.Version = "2"

			err = client.Update(updatedTestObject)

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}

			retrievedValSlice, err = client.RetrieveFromStore(updatedTestObject.AppServiceKey)

			if retrievedValSlice == nil {
				t.Fatal("No objects retrieved from store")
			}

			if !reflect.DeepEqual(retrievedValSlice[0], updatedTestObject) {
				t.Fatalf("Return value doesn't match expected.\nExpected: %v\nActual: %v\n", updatedTestObject, retrievedValSlice[0])
			}

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}

			err = client.RemoveFromStore(id)

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}

			shouldBeEmpty, err := client.RetrieveFromStore(TestAppServiceKey)

			if len(shouldBeEmpty) != 0 {
				t.Fatal("Objects should be deleted")
			}

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}
		})
	}
}
