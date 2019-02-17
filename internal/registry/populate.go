// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2019 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package registry

import (
	"fmt"
	"os"
	"reflect"
	"strconv"

	consulapi "github.com/hashicorp/consul/api"
)

func populateValue(key string, value reflect.Value, consul *consulapi.Client) error {
	var err error
	for i := 0; i < value.NumField(); i++ {
		fieldName := value.Type().Field(i).Name
		fieldValue := value.Field(i)
		keyPath := key + "/" + fieldName
		switch fieldValue.Kind() {
		case reflect.Slice:
			for j := 0; j < fieldValue.Len(); j++ {
				if err = populateSliceValue(keyPath, fieldValue.Index(j), consul); err != nil {
					break
				}
			}
		case reflect.Map:
			mapKeys := fieldValue.MapKeys()
			for _, k := range mapKeys {
				if err = populateMapValue(keyPath+"/"+k.String(), fieldValue.MapIndex(k), consul); err != nil {
					break
				}
			}
		case reflect.Struct:
			err = populateValue(keyPath, fieldValue, consul)
		case reflect.String:
			err = populateKVPair(keyPath, fieldValue.String(), consul)
		case reflect.Int:
			err = populateKVPair(keyPath, strconv.FormatInt(fieldValue.Int(), 10), consul)
		case reflect.Bool:
			err = populateKVPair(keyPath, strconv.FormatBool(fieldValue.Bool()), consul)
		default:
			errMsg := fmt.Sprintf("Unexpected fieldValue: %v for key: %s", fieldValue, keyPath)
			fmt.Println(errMsg)
			err = fmt.Errorf(errMsg)
		}
	}
	return err
}

func populateSliceValue(key string, value reflect.Value, consul *consulapi.Client) error {
	var err error
	switch value.Kind() {
	case reflect.Struct:
		//valueName := value.FieldByName(common.NameField).String()
		//if valueName != "" {
		//	err = populateValue(key+"/"+valueName, value, consul)
		//} else {
		//	err = populateValue(key, value, consul)
		//}
		// ignore struct in slice
		return nil
	case reflect.String:
		err = populateKVPair(key, value.String(), consul)
	case reflect.Int:
		err = populateKVPair(key, strconv.FormatInt(value.Int(), 10), consul)
	case reflect.Bool:
		err = populateKVPair(key, strconv.FormatBool(value.Bool()), consul)
	case reflect.Slice:
		for j := 0; j < value.Len(); j++ {
			if err = populateSliceValue(key, value.Index(j), consul); err != nil {
				break
			}
		}
	case reflect.Map:
		mapKeys := value.MapKeys()
		for _, k := range mapKeys {
			if err = populateMapValue(key, value.MapIndex(k), consul); err != nil {
				break
			}
		}
	default:
		err = populateValue(key, value, consul)
	}
	return err
}

func populateMapValue(key string, value reflect.Value, consul *consulapi.Client) error {
	var err error
	switch value.Kind() {
	case reflect.String:
		err = populateKVPair(key, value.String(), consul)
	case reflect.Int:
		err = populateKVPair(key, strconv.FormatInt(value.Int(), 10), consul)
	case reflect.Bool:
		err = populateKVPair(key, strconv.FormatBool(value.Bool()), consul)
	case reflect.Slice:
		for j := 0; j < value.Len(); j++ {
			if err = populateSliceValue(key, value.Index(j), consul); err != nil {
				break
			}
		}
	case reflect.Map:
		mapKeys := value.MapKeys()
		for _, k := range mapKeys {
			if err = populateMapValue(key, value.MapIndex(k), consul); err != nil {
				break
			}
		}
	default:
		err = populateValue(key, value, consul)
	}
	return err
}

func populateKVPair(key string, value string, consul *consulapi.Client) error {
	fmt.Fprintf(os.Stdout, "populating %v to %s \n", value, key)
	if value == "" {
		return nil
	}

	// Get a handle to the KV API
	kv := consul.KV()

	// PUT a new KV pair
	p := &consulapi.KVPair{Key: key, Value: []byte(value)}
	_, err := kv.Put(p, nil)
	return err
}
