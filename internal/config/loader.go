// -*- mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2017-2018 Canonical Ltd
// Copyright (C) 2018-2019 IOTech Ltd
// Copyright (c) 2019 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0

package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"strconv"
	"syscall"
	"time"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/go-mod-registry/pkg/types"
	"github.com/edgexfoundry/go-mod-registry/registry"
	"github.com/pelletier/go-toml"
)

var (
	RegistryClient registry.Client
)

// LoadConfig loads the local configuration file based upon the
// specified parameters and returns a pointer to the global Config
// struct which holds all of the local configuration settings for
// the DS. The string useRegisty indicates whether the registry
// should be used to read config settings, and it might contain the registry path.
// This also controls whether the service registers itself the registry.
// The profile and confDir are used to locate the local TOML config file.
func LoadConfig(useRegistry string, profile string, confDir string) (configuration *common.Config, err error) {
	fmt.Fprintf(os.Stdout, "Init: useRegistry: %v profile: %s confDir: %s\n",
		useRegistry, profile, confDir)

	var registryMsg string
	if useRegistry != "" {
		configuration = &common.Config{}
		err = parseRegistryPath(useRegistry, configuration)
		if err != nil {
			return
		}

		// TODO: Verify this is correct.
		stem := common.ConfigRegistryStem + common.ServiceName + "/"

		registryMsg = "Register in registry..."
		registryConfig := types.Config{
			Host:       configuration.Registry.Host,
			Port:       configuration.Registry.Port,
			Type:       configuration.Registry.Type,
			Stem:       stem,
			CheckRoute: common.APIPingRoute,
			ServiceKey: common.ServiceName,
		}

		RegistryClient, err = registry.NewRegistryClient(registryConfig)
		if err != nil {
			return nil, fmt.Errorf("connection to Registry could not be made: %v", err.Error())
		}

		// Check if registry service is running
		if err := checkRegistryUp(configuration); err != nil {
			return nil, err
		}

		hasConfiguration, err := RegistryClient.HasConfiguration()
		if err != nil {
			return nil, fmt.Errorf("could not verify that Registry already has configuration: %v", err.Error())
		}

		if hasConfiguration {
			// Get the configuration values from the Registry
			rawConfig, err := RegistryClient.GetConfiguration(configuration)
			if err != nil {
				return nil, fmt.Errorf("could not get configuration from Registry: %v", err.Error())
			}

			actual, ok := rawConfig.(*common.Config)
			if !ok {
				return nil, fmt.Errorf("configuration from Registry failed type check")
			}

			configuration = actual
		} else {
			// Self bootstrap the Registry with the device service's configuration
			fmt.Fprintln(os.Stdout, "Pushing configuration into Registry...")

			configuration, err = loadConfigFromFile(profile, confDir)
			if err != nil {
				return nil, err
			}

			err := RegistryClient.PutConfiguration(*configuration, true)
			if err != nil {
				return nil, fmt.Errorf("could not push configuration to Registry: %v", err.Error())
			}
		}

		// recreate a registry client because the CheckInterval and Service related config should be loaded from
		// the registry before self-registration
		registryConfig = types.Config{
			Host:          configuration.Registry.Host,
			Port:          configuration.Registry.Port,
			Type:          configuration.Registry.Type,
			Stem:          stem,
			CheckInterval: configuration.Registry.CheckInterval,
			CheckRoute:    common.APIPingRoute,
			ServiceKey:    common.ServiceName,
			ServiceHost:   configuration.Service.Host,
			ServicePort:   configuration.Service.Port,
		}

		RegistryClient, err = registry.NewRegistryClient(registryConfig)
		if err != nil {
			return nil, fmt.Errorf("connection to Registry could not be made: %v", err.Error())
		}

		// Register the service with Registry for discovery and health checks
		err = RegistryClient.Register()
		if err != nil {
			return nil, fmt.Errorf("could not register service with Registry: %v", err.Error())
		}
	} else {
		registryMsg = "Bypassing registration in registry..."
		configuration, err = loadConfigFromFile(profile, confDir)
		if err != nil {
			return nil, err
		}
	}

	fmt.Println(registryMsg)
	return configuration, nil
}

// parseRegistryPath parses the useRegistry from flag to set the url of registry, if it is valid.
// if it is invalid, return an error don't override anything by ending this func.
func parseRegistryPath(registryUrl string, config *common.Config) error {
	u, err := url.Parse(registryUrl)
	if err != nil {
		fmt.Fprintln(os.Stdout, "The format of Registry path from argument is wrong: ", err.Error())
		return err
	}

	port, err := strconv.Atoi(u.Port())
	if err != nil {
		fmt.Fprintln(os.Stdout, "The port format of Registry path from argument is wrong: ", err.Error())
		return err
	}

	config.Registry.Type = u.Scheme
	config.Registry.Host = u.Hostname()
	config.Registry.Port = port
	return nil
}

func loadConfigFromFile(profile string, confDir string) (config *common.Config, err error) {
	if len(confDir) == 0 {
		confDir = common.ConfigDirectory
	}

	if len(profile) > 0 {
		confDir = confDir + "/" + profile
	}

	path := path.Join(confDir, common.ConfigFileName)
	absPath, err := filepath.Abs(path)
	if err != nil {
		err = fmt.Errorf("Could not create absolute path to load configuration: %s; %v", path, err.Error())
		return nil, err
	}
	fmt.Fprintln(os.Stdout, fmt.Sprintf("Loading configuration from: %s\n", absPath))

	// As the toml package can panic if TOML is invalid,
	// or elements are found that don't match members of
	// the given struct, use a defered func to recover
	// from the panic and output a useful error.
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("could not load configuration file; invalid TOML (%s)", path)
		}
	}()

	config = &common.Config{}
	contents, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("Could not load configuration file (%s): %v\nBe sure to change to program folder or set working directory.", path, err.Error())
	}

	// Decode the configuration from TOML
	//
	// TODO: invalid input can cause a SIGSEGV fatal error (INVESTIGATE)!!!
	//       - test missing keys, keys with wrong type, ...
	err = toml.Unmarshal(contents, config)
	if err != nil {
		return nil, fmt.Errorf("unable to parse configuration file (%s): %v", path, err.Error())
	}

	return config, nil
}

func checkRegistryUp(config *common.Config) error {
	registryUrl := common.BuildAddr(config.Registry.Host, strconv.Itoa(config.Registry.Port))
	fmt.Println("Check registry is up...", registryUrl)
	config.Registry.FailLimit = common.RegistryFailLimit
	fails := 0
	for fails < config.Registry.FailLimit {
		if RegistryClient.IsAlive() {
			break
		}

		time.Sleep(time.Second * time.Duration(config.Registry.FailWaitTime))
		fails++
	}

	if fails >= config.Registry.FailLimit {
		return errors.New("can't get connection to Registry")
	}
	return nil
}

func ListenForConfigChanges() {
	if RegistryClient == nil {
		common.LoggingClient.Error("listenForConfigChanges() registry client not set")
		return
	}

	common.LoggingClient.Info("listen for config changes from Registry")

	errChannel := make(chan error)
	updateChannel := make(chan interface{})

	RegistryClient.WatchForChanges(updateChannel, errChannel, &common.WritableInfo{}, common.WritableKey)

	signalChan := make(chan os.Signal)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case <-signalChan:
			// Quietly and gracefully stop when SIGINT/SIGTERM received
			return
		case ex := <-errChannel:
			common.LoggingClient.Error(ex.Error())
		case raw, ok := <-updateChannel:
			if ok {
				actual, ok := raw.(*common.WritableInfo)
				if !ok {
					common.LoggingClient.Error("listenForConfigChanges() type check failed")
				}
				common.CurrentConfig.Writable = *actual
				common.LoggingClient.Info("Writeable configuration has been updated. Setting log level to " + common.CurrentConfig.Writable.LogLevel)
				common.LoggingClient.SetLogLevel(common.CurrentConfig.Writable.LogLevel)
			} else {
				return
			}
		}
	}
}
