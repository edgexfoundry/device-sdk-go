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
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/db/interfaces"
	"github.com/edgexfoundry/app-functions-sdk-go/internal/store/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Client provides a wrapper for Mongo's Client type
type Client struct {
	Timeout time.Duration
	Client  *mongo.Database
	// unexported Client for Disconnect
	client *mongo.Client
}

var mongoCollection = "store"

// Store persists a stored object to the data store.
func (c Client) Store(o models.StoredObject) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	_, err := c.Client.Collection(mongoCollection).InsertOne(ctx, &o)
	if err != nil {
		return err
	}

	return nil
}

// RetrieveFromStore gets an object from the data store.
func (c Client) RetrieveFromStore(appServiceKey string) (objects []models.StoredObject, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	filter := bson.M{"appServiceKey": appServiceKey}

	// find all documents
	cursor, err := c.Client.Collection(mongoCollection).Find(ctx, filter)
	if err != nil {
		return nil, err
	}

	// iterate through all documents
	for cursor.Next(ctx) {
		var p models.StoredObject
		// decode the document
		if err = cursor.Decode(&p); err != nil {
			return nil, err
		}
		objects = append(objects, p)
	}

	// check if the cursor encountered any errors while iterating
	if err = cursor.Err(); err != nil {
		return nil, err
	}

	return objects, nil
}

// Update replaces the data currently in the store with the provided data.
func (c Client) Update(o models.StoredObject) error {
	// set filters and updates
	filter := bson.M{"id": o.ID}

	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	// update document
	updated, err := c.Client.Collection(mongoCollection).ReplaceOne(ctx, filter, &o)
	if err != nil {
		return err
	}

	if updated.ModifiedCount == 0 {
		return errors.New("no updates performed")
	}

	return nil
}

// UpdateRetryCount modifies the RetryCount variable for a given object.
func (c Client) UpdateRetryCount(id string, count int) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	filter := bson.M{"id": id}
	update := bson.M{"$set": bson.M{"retryCount": count}}

	// update document
	updated, err := c.Client.Collection(mongoCollection).UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if updated.ModifiedCount == 0 {
		return errors.New("no updates performed")
	}

	return nil
}

// RemoveFromStore removes an object from the data store.
func (c Client) RemoveFromStore(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	deleted, err := c.Client.Collection(mongoCollection).DeleteOne(ctx, bson.M{"id": id})
	if err != nil {
		return err
	}

	if deleted.DeletedCount == 0 {
		return errors.New("no deletes performed")
	}

	return nil
}

func (c Client) Disconnect() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.Timeout)
	defer cancel()

	return c.client.Disconnect(ctx)
}

// NewClient provides a factory for building a StoreClient
func NewClient(config db.Configuration) (client interfaces.StoreClient, err error) {
	var uri string
	if config.Username == "" && config.Password == "" {
		// no auth path
		uri = fmt.Sprintf("mongodb://%s:%s", config.Host, strconv.Itoa(config.Port))
	} else {
		uri = fmt.Sprintf("mongodb://%s:%s@%s:%s", config.Username, config.Password, config.Host, strconv.Itoa(config.Port))
	}

	timeout := time.Duration(config.Timeout) * time.Millisecond
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	notify := make(chan struct{})

	var mongoDatabase *mongo.Database
	var mongoClient *mongo.Client
	go func() {
		mongoClient, err = mongo.Connect(ctx, options.Client().ApplyURI(uri))
		if err != nil {
			cancel()
			return
		}

		// check the connection
		err = mongoClient.Ping(ctx, nil)
		if err != nil {
			cancel()
			return
		}

		mongoDatabase = mongoClient.Database(config.DatabaseName)

		// ping the watcher and tell it we're done
		notify <- struct{}{}
	}()

	select {
	case <-ctx.Done():
		if err != nil {
			// an error from the business logic
			return
		} else {
			// timeout exceeded
			return nil, ctx.Err()
		}
	case <-notify:
		return Client{timeout, mongoDatabase, mongoClient}, nil
	}
}
