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

	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/contracts"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db"

	"github.com/google/uuid"
)

const (
	TestHost         = "localhost"
	TestPort         = 27017
	TestTimeout      = 5000
	TestDatabaseName = "test"
	TestBatchSize    = 1337

	TestRetryCount       = 100
	TestPipelinePosition = 1337
	TestVersion          = "your"
	TestCorrelationID    = "test"
	TestEventID          = "probably"
	TestEventChecksum    = "failed :("
)

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

var TestContract = contracts.StoredObject{
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
	EventID:          TestEventID,
	EventChecksum:    TestEventChecksum,
}

var TestContractBadID = contracts.StoredObject{
	ID:               "brandon!",
	AppServiceKey:    "brandon!",
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
	EventID:          TestEventID,
	EventChecksum:    TestEventChecksum,
}

func TestClient_NewClient(t *testing.T) {
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
			client, connectErr := NewClient(test.config)

			if client != nil {
				disconnectErr := client.Disconnect()
				if test.expectedError && disconnectErr == nil {
					t.Fatal("Expected an error")
				}

				if !test.expectedError && disconnectErr != nil {
					t.Fatalf("Unexpectedly encountered error: %s", disconnectErr.Error())
				}
			}

			if test.expectedError && connectErr == nil {
				t.Fatal("Expected an error")
			}

			if !test.expectedError && connectErr != nil {
				t.Fatalf("Unexpectedly encountered error: %s", connectErr.Error())
			}
		})
	}
}

func TestClient_Store(t *testing.T) {
	TestContractUUID := TestContract
	TestContractUUID.ID = uuid.New().String()
	TestContractUUID.AppServiceKey = uuid.New().String()

	TestContractValid := TestContract
	TestContractValid.AppServiceKey = uuid.New().String()

	TestContractNoAppServiceKey := TestContract
	TestContractUUID.ID = uuid.New().String()

	client, _ := NewClient(TestValidNoAuthConfig)

	tests := []struct {
		name          string
		toStore       contracts.StoredObject
		expectedError bool
	}{
		{
			"Success, no ID",
			TestContractValid,
			false,
		},
		{
			"Success, no ID double store",
			TestContractValid,
			false,
		},
		{
			"Success, no app service key",
			TestContractNoAppServiceKey,
			false,
		},
		{
			"Failure, no app service key double store",
			TestContractNoAppServiceKey,
			false,
		},
		{
			"Success, object with UUID",
			TestContractUUID,
			false,
		},
		{
			"Failure, object with UUID double store",
			TestContractUUID,
			true,
		},
		{
			"Bad ID",
			TestContractBadID,
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := client.Store(test.toStore)
			if test.expectedError && err == nil {
				t.Fatal("Expected an error")
			}

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}
		})
	}

	err := client.Disconnect()
	if err != nil {
		t.Fatalf("Unexpectedly encountered error: %s", err.Error())
	}
}

func TestClient_RetrieveFromStore(t *testing.T) {
	UUIDAppServiceKey := uuid.New().String()

	UUIDContract0 := TestContract
	UUIDContract0.ID = uuid.New().String()
	UUIDContract0.AppServiceKey = UUIDAppServiceKey

	UUIDContract1 := TestContract
	UUIDContract1.ID = uuid.New().String()
	UUIDContract1.AppServiceKey = UUIDAppServiceKey

	client, _ := NewClient(TestValidNoAuthConfig)

	tests := []struct {
		name          string
		toStore       []contracts.StoredObject
		key           string
		expectedError bool
	}{
		{
			"Success, single object",
			[]contracts.StoredObject{UUIDContract0},
			UUIDAppServiceKey,
			false,
		},
		{
			"Success, multiple object",
			[]contracts.StoredObject{UUIDContract0, UUIDContract1},
			UUIDAppServiceKey,
			false,
		},
		{
			"Failure, no app service key",
			[]contracts.StoredObject{},
			"",
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			for _, object := range test.toStore {
				_, _ = client.Store(object)
			}

			actual, err := client.RetrieveFromStore(test.key)

			if test.expectedError && err == nil {
				t.Fatal("Expected an error")
			}

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}

			if len(actual) != len(test.toStore) {
				t.Fatalf("Returned slice length doesn't match expected.\nExpected: %v\nActual: %v\n", len(test.toStore), len(actual))
			}
		})
	}

	err := client.Disconnect()
	if err != nil {
		t.Fatalf("Unexpectedly encountered error: %s", err.Error())
	}
}

func TestClient_Update(t *testing.T) {
	TestContractValid := TestContract
	TestContractValid.AppServiceKey = uuid.New().String()
	TestContractValid.Version = uuid.New().String()

	client, _ := NewClient(TestValidNoAuthConfig)

	// add the objects we're going to update in the database now so we have a known state
	TestContractValid.ID, _ = client.Store(TestContractValid)

	tests := []struct {
		name          string
		expectedVal   contracts.StoredObject
		expectedError bool
	}{
		{
			"Success",
			TestContractValid,
			false,
		},
		{
			"Failure, no UUID",
			TestContract,
			true,
		},
		// cannot test for no AppServiceKey since this is a valid update but an invalid read
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := client.Update(test.expectedVal)

			if test.expectedError && err == nil {
				t.Fatal("Expected an error")
			}

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}

			// only do a lookup on tests that we aren't expecting errors
			if !test.expectedError {
				actual, _ := client.RetrieveFromStore(test.expectedVal.AppServiceKey)
				if actual == nil {
					t.Fatal("No objects retrieved from store")
				}

				if !reflect.DeepEqual(actual[0], test.expectedVal) {
					t.Fatalf("Return value doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedVal, actual[0])
				}
			}
		})
	}

	err := client.Disconnect()
	if err != nil {
		t.Fatalf("Unexpectedly encountered error: %s", err.Error())
	}
}

func TestClient_RemoveFromStore(t *testing.T) {
	TestContractValid := TestContract
	TestContractValid.AppServiceKey = uuid.New().String()

	client, _ := NewClient(TestValidNoAuthConfig)

	// add the objects we're going to update in the database now so we have a known state
	TestContractValid.ID, _ = client.Store(TestContractValid)

	tests := []struct {
		name          string
		testObject    contracts.StoredObject
		expectedError bool
	}{
		{
			"Success",
			TestContractValid,
			false,
		},
		{
			"Failure, no UUID",
			TestContract,
			true,
		},
		// cannot test for no AppServiceKey since this is a valid delete but an invalid read
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := client.RemoveFromStore(test.testObject)

			if test.expectedError && err == nil {
				t.Fatal("Expected an error")
			}

			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %s", err.Error())
			}

			// only do a lookup on tests that we aren't expecting errors
			if !test.expectedError {
				actual, _ := client.RetrieveFromStore(test.testObject.AppServiceKey)
				if actual != nil {
					t.Fatal("Object retrieved, should have been nil")
				}
			}
		})
	}

	err := client.Disconnect()
	if err != nil {
		t.Fatalf("Unexpectedly encountered error: %s", err.Error())
	}
}
