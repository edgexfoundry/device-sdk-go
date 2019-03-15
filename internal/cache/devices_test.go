package cache

import (
	"context"
	"testing"

	"github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/mock"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

var ds []models.Device

func init() {
	common.ServiceName = "device-cache-test"
	common.DeviceClient = &mock.DeviceClientMock{}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	ds, _ = common.DeviceClient.DevicesForServiceByName(common.ServiceName, ctx)
}

func TestNewDeviceCache(t *testing.T) {
	dc := newDeviceCache([]models.Device{})
	if _, ok := dc.(DeviceCache); !ok {
		t.Error("the newDeviceCache function supposed to return a value which holds the DeviceCache type")
	}
}

func TestDeviceCache_ForName(t *testing.T) {
	dc := newDeviceCache(ds)

	if d, found := dc.ForName(mock.ValidDeviceRandomBoolGenerator.Name); !found {
		t.Error("supposed to find a matching device in cache by a valid device name")
	} else {
		assert.Equal(t, mock.ValidDeviceRandomBoolGenerator, d)
	}

	if _, found := dc.ForName(mock.NewValidDevice.Name); found {
		t.Error("not supposed to find a matching device in cache by an invalid device name")
	}
}

func TestDeviceCache_ForId(t *testing.T) {
	dc := newDeviceCache(ds)

	if d, found := dc.ForId(mock.ValidDeviceRandomBoolGenerator.Id); !found {
		t.Error("supposed to find a matching device in cache by a valid device id")
	} else {
		assert.Equal(t, mock.ValidDeviceRandomBoolGenerator, d)
	}

	if _, found := dc.ForId(mock.NewValidDevice.Id); found {
		t.Error("not supposed to find a matching device in cache by an invalid device id")
	}
}

func TestDeviceCache_All(t *testing.T) {
	dc := newDeviceCache(ds)
	dsFromCache := dc.All()

	for _, dFromCache := range dsFromCache {
		for _, d := range ds {
			if dFromCache.Id == d.Id {
				assert.Equal(t, d, dFromCache)
				continue
			}
		}
	}
}

func TestDeviceCache_Add(t *testing.T) {
	dc := newDeviceCache(ds)

	if err := dc.Add(mock.NewValidDevice); err != nil {
		t.Error("failed to add a new device to cache")
	}

	if d3, found := dc.ForId(mock.ValidDeviceRandomFloatGenerator.Id); !found {
		t.Error("unable to find the device which just be added to cache")
	} else {
		assert.Equal(t, mock.ValidDeviceRandomFloatGenerator, d3)
	}

	if err := dc.Add(mock.DuplicateDeviceRandomFloatGenerator); err == nil {
		t.Error("supposed to get an error when adding a duplicate device to cache")
	}
}

func TestDeviceCache_RemoveByName(t *testing.T) {
	dc := newDeviceCache(ds)

	if err := dc.RemoveByName(mock.NewValidDevice.Name); err == nil {
		t.Error("supposed to get an error when removing a device which doesn't exist in cache")
	}

	if err := dc.RemoveByName(mock.ValidDeviceRandomBoolGenerator.Name); err != nil {
		t.Error("failed to remove device from cache by name")
	}

	if _, found := dc.ForName(mock.ValidDeviceRandomBoolGenerator.Name); found {
		t.Error("unable to remove device from cache by name")
	}
}

func TestDeviceCache_Remove(t *testing.T) {
	dc := newDeviceCache(ds)

	if err := dc.Remove(mock.NewValidDevice.Id); err == nil {
		t.Error("supposed to get an error when removing a device which doesn't exist in cache")
	}

	if err := dc.Remove(mock.ValidDeviceRandomBoolGenerator.Id); err != nil {
		t.Error("failed to remove device from cache by id")
	}

	if _, found := dc.ForId(mock.ValidDeviceRandomBoolGenerator.Id); found {
		t.Error("unable to remove device from cache by id")
	}
}

func TestDeviceCache_Update(t *testing.T) {
	dc := newDeviceCache(ds)

	if err := dc.Update(mock.NewValidDevice); err == nil {
		t.Error("supposed to get an error when updating a device which doesn't exist in cache")
	}

	mock.ValidDeviceRandomBoolGenerator.AdminState = models.Locked
	if err := dc.Update(mock.ValidDeviceRandomBoolGenerator); err != nil {
		t.Error("failed to update device in cache")
	}

	if ud0, found := dc.ForId(mock.ValidDeviceRandomBoolGenerator.Id); !found {
		t.Error("unable to find the device in cache after updating it")
	} else {
		assert.Equal(t, mock.ValidDeviceRandomBoolGenerator, ud0)
	}
}

func TestDeviceCache_UpdateAdminState(t *testing.T) {
	dc := newDeviceCache(ds)

	if err := dc.UpdateAdminState(mock.NewValidDevice.Id, models.Locked); err == nil {
		t.Error("supposed to get an error when updating AdminState of the device which doesn't exist in cache")
	}
	if err := dc.UpdateAdminState(mock.ValidDeviceRandomBoolGenerator.Id, models.Locked); err != nil {
		t.Error("failed to update AdminState")
	}
	if ud0, _ := dc.ForId(mock.ValidDeviceRandomBoolGenerator.Id); ud0.AdminState != models.Locked {
		t.Error("succeeded in executing UpdateAdminState, but the value of AdminState was not updated")
	}
}
