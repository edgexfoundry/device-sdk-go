// -*- Mode: Go; indent-tabs-mode: t -*-

/*
 * Copyright (C) 2018 Canonical Ltd
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */
package cache

import (
	"sync"
	
	"github.com/edgexfoundry/core-clients-go/metadataclients"
	"github.com/edgexfoundry/core-domain-go/models"
)

type WatcherCache struct {
	devices  map[string]models.Device
	ac       metadataclients.AddressableClient
	dc       metadataclients.DeviceClient
}

var (
	wcOnce       sync.Once
	watcherCache *WatcherCache
)

// Create a singleton WatcherCache instance
func NewWatcherCache() *WatcherCache {

	wcOnce.Do(func() {
		watcherCache = &WatcherCache{}
	})

	return watcherCache
}


