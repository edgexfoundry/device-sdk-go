// -*- Mode: Go; indent-tabs-mode: t -*-
//
// # Copyright (C) 2023 Intel Corporation
//
// SPDX-License-Identifier: Apache-2.0
package provision

import (
	"context"
	goErrors "errors"
	"github.com/edgexfoundry/device-sdk-go/v4/internal/cache"
	"github.com/edgexfoundry/go-mod-bootstrap/v4/bootstrap/interfaces"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/clients/logger"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/requests"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/dtos/responses"
	"github.com/edgexfoundry/go-mod-core-contracts/v4/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"path"
	"testing"
)

var simpleProfile = responses.DeviceProfileResponse{
	Profile: dtos.DeviceProfile{
		DeviceProfileBasicInfo: dtos.DeviceProfileBasicInfo{
			Name:         "Simple-Device",
			Manufacturer: "Simple Corp.",
			Description:  "Example of Simple Device",
			Model:        "SP-01",
			Labels:       []string{"modbus"},
		},
		DeviceResources: []dtos.DeviceResource{
			{
				Description: "Switch On/Off.",
				Name:        "SwitchButton",
				IsHidden:    false,
				Properties: dtos.ResourceProperties{
					ValueType:    "Bool",
					ReadWrite:    "RW",
					DefaultValue: "true",
				},
			},
			{
				Description: "Visual representation of Switch state.",
				Name:        "Image",
				IsHidden:    false,
				Properties: dtos.ResourceProperties{
					ValueType: "Binary",
					ReadWrite: "R",
					MediaType: "image/jpeg",
				},
			},
			{
				Description: "X axis rotation rate",
				Name:        "Xrotation",
				IsHidden:    true,
				Properties: dtos.ResourceProperties{
					ValueType: "Int32",
					ReadWrite: "RW",
					Units:     "rpm",
				},
			},
			{
				Description: "y axis rotation rate",
				Name:        "yrotation",
				IsHidden:    true,
				Properties: dtos.ResourceProperties{
					ValueType: "Int32",
					ReadWrite: "RW",
					Units:     "rpm",
				},
			},
			{
				Description: "z axis rotation rate",
				Name:        "zrotation",
				IsHidden:    true,
				Properties: dtos.ResourceProperties{
					ValueType: "Int32",
					ReadWrite: "RW",
					Units:     "rpm",
				},
			},
			{
				Description: "String array",
				Name:        "StringArray",
				IsHidden:    false,
				Properties: dtos.ResourceProperties{
					ValueType: "StringArray",
					ReadWrite: "RW",
				},
			},
			{
				Description: "Unsigned 8bit array",
				Name:        "Uint8Array",
				IsHidden:    false,
				Properties: dtos.ResourceProperties{
					ValueType: "Uint8Array",
					ReadWrite: "RW",
				},
			},
			{
				Description: "Counter data",
				Name:        "Counter",
				IsHidden:    false,
				Properties: dtos.ResourceProperties{
					ValueType: "Object",
					ReadWrite: "RW",
				},
			},
		},
		DeviceCommands: []dtos.DeviceCommand{
			{
				Name:      "Switch",
				IsHidden:  false,
				ReadWrite: "RW",
				ResourceOperations: []dtos.ResourceOperation{
					{
						DeviceResource: "SwitchButton",
						DefaultValue:   "false",
					},
				},
			},
			{
				Name:      "Image",
				IsHidden:  false,
				ReadWrite: "R",
				ResourceOperations: []dtos.ResourceOperation{
					{
						DeviceResource: "Image",
					},
				},
			},
			{
				Name:      "Rotation",
				IsHidden:  false,
				ReadWrite: "RW",
				ResourceOperations: []dtos.ResourceOperation{
					{
						DeviceResource: "XRotation",
						DefaultValue:   "0",
					},
					{
						DeviceResource: "YRotation",
						DefaultValue:   "0",
					},
					{
						DeviceResource: "ZRotation",
						DefaultValue:   "0",
					},
				},
			},
		},
	},
}

func Test_processProfiles(t *testing.T) {
	tests := []struct {
		name                string
		path                string
		secretProvider      interfaces.SecretProvider
		profileName         string
		dpcMockRes          responses.DeviceProfileResponse
		dpcMockErr          errors.EdgeX
		expectedNumProfiles int
		expectedEdgexErrMsg string
	}{
		{"valid load profile from file, profile exists", path.Join("..", "..", "example", "cmd", "device-simple", "res", "profiles", "Simple-Driver.yaml"), nil, "Simple-Device", simpleProfile, nil, 0, ""},
		{"valid load profile from file, profile does not exist in metadata", path.Join("..", "..", "example", "cmd", "device-simple", "res", "profiles", "Simple-Driver.yaml"), nil, "Simple-Device", responses.DeviceProfileResponse{}, errors.NewCommonEdgeXWrapper(goErrors.New("could not find profile")), 1, ""},
		{"valid load profile from uri, profile exists", "https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/example/cmd/device-simple/res/profiles/Simple-Driver.yaml", nil, "Simple-Device", simpleProfile, nil, 0, ""},
		{"valid load profile from uri, profile does not exist in metadata", "https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/example/cmd/device-simple/res/profiles/Simple-Driver.yaml", nil, "Simple-Device", responses.DeviceProfileResponse{}, errors.NewCommonEdgeXWrapper(goErrors.New("could not find profile")), 1, ""},
		{"invalid load empty profile from file", "", nil, "Simple-Device", responses.DeviceProfileResponse{}, errors.NewCommonEdgeXWrapper(goErrors.New("could not find profile")), 0, ""},
		{"invalid load profile from bogus file", path.Join("..", "..", "example", "cmd", "device-simple", "res", "profiles", "bogus.yaml"), nil, "Simple-Device", responses.DeviceProfileResponse{}, nil, 0, ""},
		{"invalid load profile from bogus uri", "https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/example/cmd/device-simple/res/profiles/bogus.yaml", nil, "Simple-Device", responses.DeviceProfileResponse{}, nil, 0, ""},
		{"invalid load profile from file, duplicate profile", path.Join("..", "..", "example", "cmd", "device-simple", "res", "profiles", "Simple-Driver.yaml"), nil, "Simple-Device", profile, nil, 0, "Profile testProfile has already existed in cache"},
		{"invalid load profile from uri, duplicate profile", "https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/example/cmd/device-simple/res/profiles/Simple-Driver.yaml", nil, "Simple-Device", profile, nil, 0, "Profile testProfile has already existed in cache"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var addProfilesReq []requests.DeviceProfileRequest
			var edgexErr errors.EdgeX
			lc := logger.MockLogger{}
			dic, dpcMock := NewMockDIC()
			dpcMock.On("DeviceProfileByName", context.Background(), tt.profileName).Return(tt.dpcMockRes, tt.dpcMockErr)
			err := cache.InitCache(TestDeviceService, TestDeviceService, dic)
			require.NoError(t, err)
			addProfilesReq, edgexErr = processProfiles(tt.path, tt.path, tt.secretProvider, lc, dpcMock)
			assert.Equal(t, len(addProfilesReq), tt.expectedNumProfiles)
			if edgexErr != nil {
				assert.Contains(t, edgexErr.Error(), tt.expectedEdgexErrMsg)
			}
		})
	}
}

func Test_loadProfilesFromURI(t *testing.T) {
	tests := []struct {
		name                string
		path                string
		secretProvider      interfaces.SecretProvider
		profileNames        []string
		dpcMockRes          []responses.DeviceProfileResponse
		dpcMockErr          []errors.EdgeX
		expectedNumProfiles int
		expectedEdgexErrMsg string
	}{
		{"valid load from uri, profile exists",
			"https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/internal/provision/uri-test-files/profiles/index.json",
			nil,
			[]string{"Simple-Device", "Simple-Device2"},
			[]responses.DeviceProfileResponse{simpleProfile, simpleProfile},
			[]errors.EdgeX{nil, nil},
			0, ""},
		{"valid load from uri, profile does not exist in metadata",
			"https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/internal/provision/uri-test-files/profiles/index.json",
			nil,
			[]string{"Simple-Device", "Simple-Device2"},
			[]responses.DeviceProfileResponse{responses.DeviceProfileResponse{}, responses.DeviceProfileResponse{}},
			[]errors.EdgeX{errors.NewCommonEdgeXWrapper(goErrors.New("could not find profile")), errors.NewCommonEdgeXWrapper(goErrors.New("could not find profile"))},
			2, ""},
		{"valid load where one profile exists and one profile does not exist in metadata",
			"https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/internal/provision/uri-test-files/profiles/index.json",
			nil,
			[]string{"Simple-Device", "Simple-Device2"},
			[]responses.DeviceProfileResponse{simpleProfile, responses.DeviceProfileResponse{}},
			[]errors.EdgeX{nil, errors.NewCommonEdgeXWrapper(goErrors.New("could not find profile"))},
			1, ""},
		{"invalid load profile path join breaks",
			"https://raw.githubusercontent.com/edgexfoundry/device-sdk-go/main/internal/provision/uri-test-files/profiles/bogus.json",
			nil,
			[]string{},
			[]responses.DeviceProfileResponse{},
			[]errors.EdgeX{},
			0, "failed to load Device Profile list from URI"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var addProfilesReq []requests.DeviceProfileRequest
			var edgexErr errors.EdgeX
			lc := logger.MockLogger{}
			dic, dpcMock := NewMockDIC()
			for index := range tt.profileNames {
				dpcMock.On("DeviceProfileByName", context.Background(), tt.profileNames[index]).Return(tt.dpcMockRes[index], tt.dpcMockErr[index])
			}
			edgexErr = cache.InitCache(TestDeviceService, TestDeviceService, dic)
			require.NoError(t, edgexErr)
			parsedURI, err := url.Parse(tt.path)
			require.NoError(t, err)
			addProfilesReq, edgexErr = loadProfilesFromURI(tt.path, parsedURI, dpcMock, tt.secretProvider, lc)
			assert.Equal(t, len(addProfilesReq), tt.expectedNumProfiles)
			if edgexErr != nil {
				assert.Contains(t, edgexErr.Error(), tt.expectedEdgexErrMsg)
			}
		})
	}
}
