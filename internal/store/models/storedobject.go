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

// models describes the data types that will be used when storing export data.
package models

// StoredObject is the atomic and most abstract description of what is collected by the export store system.
type StoredObject struct {
	// id is the unique identifier for this record once stored.
	id string

	// AppServiceKey identifies the app to which this data belongs.
	AppServiceKey string

	// Payload is the data to be exported
	Payload []byte

	// RetryCount is how many times this has tried to be exported
	RetryCount int

	// PipelinePosition is where to pickup in the pipeline
	PipelinePosition int

	// Version is a hash of the functions to know if the pipeline has changed.
	Version string

	// CorrelationID is an identifier provided by EdgeX to track this record as it moves
	CorrelationID string

	// EventID is used to identify an EdgeX event from the core services and mark it as pushed.
	EventID string

	// EventChecksum is used to identify CBOR encoded data from the core services and mark it as pushed.
	EventChecksum string
}

// GetId returns the unexported field id.
func (o StoredObject) GetId() string {
	return o.id
}

// NewStoredObject creates a new instance of StoredObject and is the preferred way to create one.
func NewStoredObject(id string, appServiceKey string, payload []byte, pipelinePosition int,
	version string) StoredObject {
	return StoredObject{
		id:               id,
		AppServiceKey:    appServiceKey,
		Payload:          payload,
		RetryCount:       0,
		PipelinePosition: pipelinePosition,
		Version:          version,
	}
}
