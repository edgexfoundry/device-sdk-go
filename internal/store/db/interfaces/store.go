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

// interfaces establishes the contract required for any implementation of the export store functionality in a database provider.
package interfaces

import (
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/contracts"
)

// StoreClient establishes the contracts required to persist exported data before being forwarded.
type StoreClient interface {
	// Store persists a stored object to the data store and returns the assigned UUID.
	Store(o contracts.StoredObject) (id string, err error)

	// RetrieveFromStore gets an object from the data store.
	RetrieveFromStore(appServiceKey string) (objects []contracts.StoredObject, err error)

	// Update replaces the data currently in the store with the provided data.
	Update(o contracts.StoredObject) error

	// RemoveFromStore removes an object from the data store.
	RemoveFromStore(o contracts.StoredObject) error

	// Disconnect ends the connection.
	Disconnect() error
}
