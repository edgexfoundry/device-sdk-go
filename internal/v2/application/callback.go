// -*- Mode: Go; indent-tabs-mode: t -*-
//
// Copyright (C) 2020 IOTech Ltd
//
// SPDX-License-Identifier: Apache-2.0

package application

import (
	"fmt"

	"github.com/edgexfoundry/device-sdk-go/internal/autoevent"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/v2/cache"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/bootstrap/container"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/errors"
	contract "github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/models"
)

func AddProfile(profileRequest requests.DeviceProfileRequest, lc logger.LoggingClient) errors.EdgeX {
	err := cache.Profiles().Add(dtos.ToDeviceProfileModel(profileRequest.Profile))

	if err != nil {
		errMsg := fmt.Sprintf("failed to add profile %s", profileRequest.Profile.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	lc.Debug(fmt.Sprintf("profile %s added", profileRequest.Profile.Name))
	return nil
}

func UpdateProfile(profileRequest requests.DeviceProfileRequest, lc logger.LoggingClient) errors.EdgeX {
	_, ok := cache.Profiles().ForName(profileRequest.Profile.Name)
	if !ok {
		errMsg := fmt.Sprintf("failed to find profile %s", profileRequest.Profile.Name)
		return errors.NewCommonEdgeX(errors.KindInvalidId, errMsg, nil)
	}

	err := cache.Profiles().Update(dtos.ToDeviceProfileModel(profileRequest.Profile))
	if err != nil {
		errMsg := fmt.Sprintf("failed to update profile %s", profileRequest.Profile.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	lc.Debug(fmt.Sprintf("profile %s updated", profileRequest.Profile.Name))
	return nil
}

func DeleteProfile(id string, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	err := cache.Profiles().RemoveById(id)
	if err != nil {
		errMsg := fmt.Sprintf("failed to remove profile with given id %s", id)
		return errors.NewCommonEdgeX(errors.KindInvalidId, errMsg, err)
	}

	lc.Info(fmt.Sprintf("Removed profile with given id %s", id))
	return nil
}

func AddDevice(addDeviceRequest requests.AddDeviceRequest, dic *di.Container) errors.EdgeX {
	device := dtos.ToDeviceModel(addDeviceRequest.Device)
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	// TODO: uncomment when core-contracts v2 client is ready.
	//edgexErr := updateAssociatedProfile(device.ProfileName, dic)
	//if edgexErr != nil {
	//	errMsg := fmt.Sprintf("failed to update device profile %s", device.ProfileName)
	//	return errors.NewCommonEdgeX(errors.KindServerError, errMsg, edgexErr)
	//}

	edgexErr := cache.Devices().Add(device)
	if edgexErr != nil {
		errMsg := fmt.Sprintf("failed to add device %s", device.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, edgexErr)
	}
	lc.Debug(fmt.Sprintf("device %s added", device.Name))

	driver := container.ProtocolDriverFrom(dic.Get)
	err := driver.AddDevice(device.Name, transformDeviceProtocols(device.Protocols), contract.AdminState(device.AdminState))
	if err == nil {
		lc.Debug(fmt.Sprintf("Invoked driver.AddDevice callback for %s", device.Name))
	} else {
		errMsg := fmt.Sprintf("driver.AddDevice callback failed for %s", device.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	lc.Debug(fmt.Sprintf("Handler - starting AutoEvents for device %s", device.Name))
	autoevent.GetManager().RestartForDevice(device.Name, dic)
	return nil
}

func UpdateDevice(updateDeviceRequest requests.UpdateDeviceRequest, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)

	device, ok := cache.Devices().ForName(*updateDeviceRequest.Device.Name)
	if !ok {
		errMsg := fmt.Sprintf("failed to find device %s", *updateDeviceRequest.Device.Name)
		return errors.NewCommonEdgeX(errors.KindInvalidId, errMsg, nil)
	}

	requests.ReplaceDeviceModelFieldsWithDTO(&device, updateDeviceRequest.Device)
	// TODO: uncomment when core-contracts v2 client is ready.
	//edgexErr := updateAssociatedProfile(device.ProfileName, dic)
	//if edgexErr != nil {
	//	errMsg := fmt.Sprintf("failed to update device profile %s", device.ProfileName)
	//	return errors.NewCommonEdgeX(errors.KindServerError, errMsg, edgexErr)
	//}

	edgexErr := cache.Devices().Update(device)
	if edgexErr != nil {
		errMsg := fmt.Sprintf("failed to update device %s", device.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, edgexErr)
	}
	lc.Debug(fmt.Sprintf("device %s updated", device.Name))

	driver := container.ProtocolDriverFrom(dic.Get)
	err := driver.UpdateDevice(device.Name, transformDeviceProtocols(device.Protocols), contract.AdminState(device.AdminState))
	if err == nil {
		lc.Debug(fmt.Sprintf("Invoked driver.UpdateDevice callback for %s", device.Name))
	} else {
		errMsg := fmt.Sprintf("driver.UpdateDevice callback failed for %s", device.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	lc.Debug(fmt.Sprintf("Handler - starting AutoEvents for device %s", device.Name))
	autoevent.GetManager().RestartForDevice(device.Name, dic)
	return nil
}

func DeleteDevice(id string, dic *di.Container) errors.EdgeX {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	// check the device exist and stop its autoevents
	device, ok := cache.Devices().ForId(id)
	if ok {
		lc.Debug(fmt.Sprintf("Handler - stopping AutoEvents for device %s", device.Name))
		autoevent.GetManager().StopForDevice(device.Name)
	} else {
		errMsg := fmt.Sprintf("failed to find device with given id %s", id)
		return errors.NewCommonEdgeX(errors.KindInvalidId, errMsg, nil)
	}

	// remove the device in cache
	edgexErr := cache.Devices().RemoveById(id)
	if edgexErr != nil {
		errMsg := fmt.Sprintf("failed to remove device %s", device.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, edgexErr)
	}
	lc.Info(fmt.Sprintf("Removed device: %s", device.Name))

	driver := container.ProtocolDriverFrom(dic.Get)
	err := driver.RemoveDevice(device.Name, transformDeviceProtocols(device.Protocols))
	if err == nil {
		lc.Debug(fmt.Sprintf("Invoked driver.RemoveDevice callback for %s", device.Name))
	} else {
		errMsg := fmt.Sprintf("driver.RemoveDevice callback failed for %s", device.Name)
		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	}

	// a special case in which user updates the device profile after deleting all
	// devices in metadata, the profile won't be correctly updated because metadata
	// does not know which device service callback it needs to call. Remove the unused
	// device profile in cache so that if it is updated in metadata, next time the
	// device using it is added/updated, the cache can receive the updated one as well.
	if cache.CheckProfileNotUsed(device.ProfileName) {
		edgexErr = cache.Profiles().RemoveByName(device.ProfileName)
		if edgexErr != nil {
			lc.Warn("failed to remove unused profile", edgexErr.DebugMessages())
		}
	}

	return nil
}

// updateAssociatedProfile updates the profile specified in AddDeviceRequest or UpdateDeviceRequest
// to stay consistent with core metadata.
func updateAssociatedProfile(profileName string, dic *di.Container) errors.EdgeX {
	// TODO: uncomment when core-contracts v2 clients are ready.
	//lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	//gc := container.GeneralClientFrom(dic.Get)
	//vdc := container.CoredataValueDescriptorClientFrom(dic.Get)
	//dpc := container.MetadataDeviceProfileClientFrom(dic.Get)
	//
	//profile, err := dpc.DeviceProfileForName(context.Background(), profileName)
	//if err != nil {
	//	errMsg := fmt.Sprintf("failed to find profile %s in metadata", profileName)
	//	return errors.NewCommonEdgeX(errors.KindInvalidId, errMsg, nil)
	//}
	//
	//_, exist := cache.Profiles().ForName(profileName)
	//if exist == false {
	//	err = cache.Profiles().Add(profile)
	//	if err == nil {
	//		provision.CreateDescriptorsFromProfile(&profile, lc, gc, vdc)
	//		lc.Info(fmt.Sprintf("Added device profile: %s", profileName))
	//	} else {
	//		errMsg := fmt.Sprintf("failed to add profile %s", profileName)
	//		return errors.NewCommonEdgeX(errors.KindServerError, errMsg, err)
	//	}
	//} else {
	//	err := cache.Profiles().Update(profile)
	//	if err != nil {
	//		lc.Warn(fmt.Sprintf("failed to to update profile %s in cache, using the original one", profileName))
	//	}
	//}

	return nil
}

// TODO: remove this helper function when we fully moving to v2 API
// transformDeviceProtocols transforms device protocol from v2 model to v1 model
func transformDeviceProtocols(protocols map[string]models.ProtocolProperties) map[string]contract.ProtocolProperties {
	var res = make(map[string]contract.ProtocolProperties)
	for name, protocol := range protocols {
		var p contract.ProtocolProperties = make(map[string]string)
		for k, v := range protocol {
			p[k] = v
		}
		res[name] = p
	}

	return res
}
