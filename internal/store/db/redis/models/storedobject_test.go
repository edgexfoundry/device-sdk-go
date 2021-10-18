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

package models

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/contracts"
)

var TestUUIDValid = "fb49a277-9edf-4489-a89c-235b365107f7"
var TestPayload = []byte("brandon wrote this")
var TestContextData = map[string]string{"test": "data"}

const (
	TestAppServiceKey    = "apps"
	TestRetryCount       = 2
	TestPipelinePosition = 1337
	TestVersion          = "your"
	TestCorrelationID    = "test"
)

var TestContractValid = contracts.StoredObject{
	ID:               TestUUIDValid,
	AppServiceKey:    TestAppServiceKey,
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
	ContextData:      TestContextData,
}

var TestModelValid = StoredObject{
	ID:               TestUUIDValid,
	AppServiceKey:    TestAppServiceKey,
	Payload:          TestPayload,
	RetryCount:       TestRetryCount,
	PipelinePosition: TestPipelinePosition,
	Version:          TestVersion,
	CorrelationID:    TestCorrelationID,
	ContextData:      TestContextData,
}

var TestModelEmpty = StoredObject{}

func TestStoredObject_FromContract(t *testing.T) {
	tests := []struct {
		testName       string
		fromContract   contracts.StoredObject
		expectedResult StoredObject
	}{
		{
			"Success",
			TestContractValid,
			TestModelValid,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(tt *testing.T) {
			actual := StoredObject{}
			actual.FromContract(test.fromContract)

			if !reflect.DeepEqual(actual, test.expectedResult) {
				t.Fatalf("Return value doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedResult, actual)
			}
		})
	}
}

func TestStoredObject_ToContract(t *testing.T) {
	tests := []struct {
		testName       string
		fromModel      StoredObject
		expectedResult contracts.StoredObject
	}{
		{
			"Success, UUID",
			TestModelValid,
			TestContractValid,
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(tt *testing.T) {
			actual := test.fromModel.ToContract()

			if !reflect.DeepEqual(actual, test.expectedResult) {
				t.Fatalf("Return value doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedResult, actual)
			}
		})
	}
}

func TestStoredObject_MarshalJSON(t *testing.T) {
	tests := []struct {
		name           string
		o              StoredObject
		expectedError  bool
		expectedResult string
	}{
		{
			"Successful marshalling",
			TestModelValid,
			false,
			`{"id":"fb49a277-9edf-4489-a89c-235b365107f7","appServiceKey":"apps","payload":"YnJhbmRvbiB3cm90ZSB0aGlz","retryCount":2,"pipelinePosition":1337,"version":"your","correlationID":"test","contextData":{"test":"data"}}`,
		},
		{
			"Successful, empty",
			TestModelEmpty,
			false,
			"{}",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual, err := test.o.MarshalJSON()
			if !test.expectedError && err != nil {
				t.Fatalf("Unexpectedly encountered error: %v", err)
			}

			expected := []byte(test.expectedResult)
			if !bytes.Equal(actual, expected) {
				t.Fatalf("Return value doesn't match expected.\nExpected: %v\nActual: %v\n", test.expectedResult, string(actual))
			}
		})
	}
}

func TestStoredObject_UnmarshalJSON(t *testing.T) {
	type args struct {
		data []byte
	}
	tests := []struct {
		name     string
		expected StoredObject
		args     args
		wantErr  bool
	}{
		{
			"Valid",
			TestModelValid,
			args{[]byte(`{"id":"fb49a277-9edf-4489-a89c-235b365107f7","appServiceKey":"apps","payload":[98,114,97,110,100,111,110,32,119,114,111,116,101,32,116,104,105,115],"retryCount":2,"pipelinePosition":1337,"version":"your","correlationID":"test","eventID":"probably","eventChecksum":"failed :(","contextData":{"test":"data"}}`)},
			false,
		},
		{
			"Empty",
			TestModelEmpty,
			args{[]byte("{}")},
			false,
		},
		{
			"Invalid",
			StoredObject{},
			args{[]byte(`{"author":"brandon"}`)},
			true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := new(StoredObject)
			err := actual.UnmarshalJSON(test.args.data)
			if err != nil {
				t.Fatalf("Unexpectedly encountered error: %v", err)
			}

			if !reflect.DeepEqual(*actual, test.expected) {
				t.Fatalf("Return value doesn't match expected.\nExpected: %v\nActual: %v\n", test.expected, *actual)
			}
		})
	}
}
