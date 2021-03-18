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

// redis provides the Redis implementation of the StoreClient interface.
package redis

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/contracts"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/v2/internal/store/db/redis/models"

	bootstrapConfig "github.com/edgexfoundry/go-mod-bootstrap/v2/config"

	"github.com/gomodule/redigo/redis"
)

var currClient *Client // a singleton so Readings can be de-referenced
var once sync.Once

const nameSpace = "store"

// Client provides an implementation for the Client interface for Redis
type Client struct {
	Pool      *redis.Pool // A thread-safe pool of connections to Redis
	BatchSize int
}

// Store persists a stored object to the data store. Three ("Three shall be the number thou shalt
// count, and the number of the counting shall be three") keys are used:
// * the object id to point to a STRING which is the marshalled JSON.
// * the object AppServiceKey to point to a SET containing all object ids associated with this
//   app service. Note the key is prefixed to avoid key collisions.
// * the object id to point to a HASH which contains the object AppServiceKey.
func (c Client) Store(o contracts.StoredObject) (string, error) {
	err := o.ValidateContract(false)
	if err != nil {
		return "", err
	}

	conn := c.Pool.Get()
	defer func() { _ = conn.Close() }()

	exists, err := redis.Bool(conn.Do("EXISTS", o.ID))
	if err != nil {
		return "", err
	} else if exists {
		return "", errors.New("object exists in database")
	}

	var model models.StoredObject
	model.FromContract(o)

	json, err := model.MarshalJSON()
	if err != nil {
		return "", err
	}

	_ = conn.Send("MULTI")
	_ = conn.Send("SET", model.ID, json)
	_ = conn.Send("SADD", nameSpace+":idl:"+model.AppServiceKey, model.ID)
	_ = conn.Send("HSET", nameSpace+":ask:"+model.ID, "ASK", model.AppServiceKey)
	_, err = conn.Do("EXEC")
	if err != nil {
		return "", err
	}

	if model.ID == "" {
		return "", errors.New("no ID produced")
	}

	return model.ID, nil
}

// RetrieveFromStore gets an object from the data store.
func (c Client) RetrieveFromStore(appServiceKey string) (objects []contracts.StoredObject, err error) {
	// do not satisfy requests for a blank ASK
	if appServiceKey == "" {
		return nil, errors.New("no AppServiceKey provided")
	}

	conn := c.Pool.Get()
	defer func() { _ = conn.Close() }()

	ids, err := redis.Values(conn.Do("SMEMBERS", nameSpace+":idl:"+appServiceKey))
	if err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return nil, nil
	}

	values, err := redis.ByteSlices(conn.Do("MGET", ids...))
	if err != nil {
		return nil, err
	}

	var model models.StoredObject

	for _, bytes := range values {
		err = model.UnmarshalJSON(bytes)
		if err != nil {
			return nil, err
		}
		objects = append(objects, model.ToContract())
	}

	return objects, nil
}

// Update replaces the data currently in the store with the provided data.
func (c Client) Update(o contracts.StoredObject) error {
	err := o.ValidateContract(true)
	if err != nil {
		return err
	}

	conn := c.Pool.Get()
	defer func() { _ = conn.Close() }()

	// retrieve the current AppServiceKey for this store object
	currentASK, err := redis.String(conn.Do("HGET", nameSpace+":ask:"+o.ID, "ASK"))
	if err != nil {
		return err
	}

	_ = conn.Send("MULTI")

	// ASK has changed, update the ASK registry
	if o.AppServiceKey != currentASK {
		_ = conn.Send("SREM", nameSpace+":idl:"+currentASK, o.ID)
		_ = conn.Send("SADD", nameSpace+":idl:"+o.AppServiceKey, o.ID)
		_ = conn.Send("HSET", nameSpace+":ask:"+o.ID, "ASK", o.AppServiceKey)
	}

	var update models.StoredObject
	update.FromContract(o)
	json, err := update.MarshalJSON()
	if err != nil {
		return err
	}

	_ = conn.Send("SET", update.ID, json)

	_, err = conn.Do("EXEC")
	if err != nil {
		return err
	}

	return nil
}

// RemoveFromStore removes an object from the data store.
func (c Client) RemoveFromStore(o contracts.StoredObject) error {
	err := o.ValidateContract(true)
	if err != nil {
		return err
	}

	conn := c.Pool.Get()
	defer func() { _ = conn.Close() }()

	_ = conn.Send("MULTI")
	// remove the object's representation
	_ = conn.Send("UNLINK", o.ID)
	// remove the association with the ASK
	_ = conn.Send("SREM", nameSpace+":idl:"+o.AppServiceKey, o.ID)
	_ = conn.Send("UNLINK", nameSpace+":ask:"+o.ID)

	res, err := redis.Values(conn.Do("EXEC"))
	if err != nil {
		return err
	}
	exists, _ := redis.Bool(res[0], nil)
	if !exists {
		return errors.New("could not remove object from store")
	}

	return nil
}

// Disconnect ends the connection.
func (c Client) Disconnect() error {
	return c.Pool.Close()
}

// NewClient provides a factory for building a StoreClient
func NewClient(config db.DatabaseInfo, credentials bootstrapConfig.Credentials) (interfaces.StoreClient, error) {
	var retErr error
	once.Do(func() {
		connectionString := fmt.Sprintf("%s:%d", config.Host, config.Port)
		connectTimeout, err := time.ParseDuration(config.Timeout)
		if err != nil {
			retErr = fmt.Errorf("config.Timeout failed to parse: %v", err)
			return
		}

		opts := []redis.DialOption{
			redis.DialPassword(credentials.Password),
			redis.DialConnectTimeout(connectTimeout),
		}

		dialFunc := func() (redis.Conn, error) {
			conn, err := redis.Dial(
				"tcp", connectionString, opts...,
			)
			if err != nil {
				return nil, fmt.Errorf("Could not dial Redis: %s", err)
			}
			return conn, nil
		}
		currClient = &Client{
			Pool: &redis.Pool{
				IdleTimeout: connectTimeout,
				/* The current implementation processes nested structs using concurrent connections.
				 * With the deepest nesting level being 3, three shall be the number of maximum open
				 * idle connections in the pool, to allow reuse.
				 * TODO: Once we have a concurrent benchmark, this should be revisited.
				 * TODO: Longer term, once the objects are clean of external dependencies, the use
				 * of another serializer should make this moot.
				 */
				MaxIdle: config.MaxIdle,
				Dial:    dialFunc,
			},
			BatchSize: config.BatchSize,
		}
	})

	return currClient, retErr
}
