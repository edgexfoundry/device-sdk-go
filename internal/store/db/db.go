/*******************************************************************************
 * Copyright 2019 Dell Inc.
 * Copyright (c) 2021 Intel Corporation
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

// db provides useful constants, identifiers, and simple types that apply to all implementations of the store
package db

import "errors"

const (
	// Database providers
	RedisDB = "redisdb"
)

var (
	ErrUnsupportedDatabase = errors.New("unsupported database type")
)

type DatabaseInfo struct {
	Type    string
	Host    string
	Port    int
	Timeout string

	// Redis specific configuration items
	MaxIdle   int
	BatchSize int
}
