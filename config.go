// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
//
// SPDX-License-Identifier: Apache-2.0
//
package gxds

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	ServiceName                  string
	ServiceHost                  string
	ServicePort                  int
	Labels                       []string
	Timeout                      int
	OpenMessage                  string
	ConnectRetries               int
	ConnectWait                  int
	ConnectInterval              int
	MaxLimit                     int
	HeartBeatTime                int
	DataTransform                bool
	MetadataHost                 string
	MetadataPort                 int
	DataHost                     string
	DataPort                     int
	LoggingFile                  string
	LoggingRemoteURL             string
	DefaultScheduleName          string
	DefaultScheduleFrequency     string
	DefaultScheduleEventName     string
	DefaultScheduleEventPath     string
	DefaultScheduleEventService  string
	DefaultScheduleEventSchedule string
}

func LoadConfig(configPath *string) (config *Config, err error) {
	config = &Config{}
	f, err := os.Open(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config file: %s; open failed: %v\n", *configPath, err)
		return config, err
	}
	defer f.Close()

	fmt.Fprintf(os.Stdout, "config file opened: %s\n", *configPath)

	jsonParser := json.NewDecoder(f)
	err = jsonParser.Decode(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return config, err
	}

	fmt.Fprintf(os.Stdout, "name: %v\n", config)

	return config, nil
}
