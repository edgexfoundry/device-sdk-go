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

// mongo provides the Mongo implementation of the StoreClient interface.
package mongo

import (
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/models"
)

// Store persists a stored object to the data store.
func Store(o models.StoredObject) error {
	return nil
}

// RetrieveFromStore gets an object from the data store.
func RetrieveFromStore(appServiceKey string) (objects []models.StoredObject, err error) {
	return nil, nil
}

// Update replaces the data currently in the store with the provided data.
func Update(o models.StoredObject) error {
	return nil
}

// UpdateRetryCount modifies the RetryCount variable for a given object.
func UpdateRetryCount(id string, count int) error {
	return nil
}

// RemoveFromStore removes an object from the data store.
func RemoveFromStore(id string) error {
	return nil
}

// NewClient provides a factory for building a StoreClient
func NewClient(config db.Configuration) (interfaces.StoreClient, error) {
	return nil, nil
}
