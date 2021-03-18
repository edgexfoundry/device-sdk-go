/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright 2021 Intel Inc.
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

package store

import (
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db/redis"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"
)

func NewStoreClient(config db.DatabaseInfo, credentials bootstrapConfig.Credentials) (interfaces.StoreClient, error) {
	switch config.Type {
	case db.RedisDB:
		return redis.NewClient(config, credentials)
	default:
		return nil, db.ErrUnsupportedDatabase
	}
}
