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

var vs []models.ValueDescriptor

func init() {
	common.ValueDescriptorClient = &mock.ValueDescriptorMock{}
	ctx := context.WithValue(context.Background(), common.CorrelationHeader, uuid.New().String())
	vs, _ = common.ValueDescriptorClient.ValueDescriptors(ctx)
}

func TestValueDescriptorCache(t *testing.T) {
	dpc := newValueDescriptorCache([]models.ValueDescriptor{})
	if _, ok := dpc.(ValueDescriptorCache); !ok {
		t.Error("the newValueDescriptorCache function supposed to return a value which holds the ValueDescriptorCache type")
	}
}

func TestValueDescriptorCache_ForName(t *testing.T) {
	vc := newValueDescriptorCache(vs)
	if v0, ok := vc.ForName(mock.ValueDescriptorBool.Name); !ok {
		t.Error("supposed to find a matching value descriptor in cache by a valid name")
	} else {
		assert.Equal(t, mock.ValueDescriptorBool, v0)
	}
}

func TestValueDescriptorCache_All(t *testing.T) {
	vc := newValueDescriptorCache(vs)
	vsFromCache := vc.All()

	for _, vFromCache := range vsFromCache {
		for _, v := range vs {
			if vFromCache.Id == v.Id {
				assert.Equal(t, v, vFromCache)
				continue
			}
		}
	}
}

func TestValueDescriptorCache_Add(t *testing.T) {
	vc := newValueDescriptorCache(vs)

	if err := vc.Add(mock.NewValueDescriptor); err != nil {
		t.Error("failed to add a new value descriptor to cache")
	}
	if v, found := vc.ForName(mock.NewValueDescriptor.Name); !found {
		t.Error("unable to find the value descriptor which just be added to cache")
	} else {
		assert.Equal(t, mock.NewValueDescriptor, v)
	}
	if err := vc.Add(mock.DuplicateValueDescriptorInt16); err == nil {
		t.Error("supposed to get an error when adding a duplicate value descriptor to cache")
	}
}

func TestValueDescriptorCache_RemoveByName(t *testing.T) {
	vc := newValueDescriptorCache(vs)

	if err := vc.RemoveByName(mock.NewValueDescriptor.Name); err == nil {
		t.Error("supposed to get an error when removing a value descriptor which doesn't exist in cache")
	}

	if err := vc.RemoveByName(mock.ValueDescriptorBool.Name); err != nil {
		t.Error("failed to remove value descriptor from cache by name")
	}

	if _, found := vc.ForName(mock.ValueDescriptorBool.Name); found {
		t.Error("unable to remove value descriptor from cache by name")
	}
}

func TestValueDescriptorCache_Remove(t *testing.T) {
	vc := newValueDescriptorCache(vs)

	if err := vc.Remove(mock.NewValueDescriptor.Id); err == nil {
		t.Error("supposed to get an error when removing a value descriptor which doesn't exist in cache")
	}

	if err := vc.Remove(mock.ValueDescriptorBool.Id); err != nil {
		t.Error("failed to remove value descriptor from cache by id")
	}

	if _, found := vc.ForName(mock.ValueDescriptorBool.Name); found {
		t.Error("unable to remove value descriptor from cache by id")
	}
}

func TestValueDescriptorCache_Update(t *testing.T) {
	vc := newValueDescriptorCache(vs)

	if err := vc.Update(mock.NewValueDescriptor); err == nil {
		t.Error("supposed to get an error when updating a value descriptor which doesn't exist in cache")
	}

	mock.DeviceProfileRandomBoolGenerator.Description = "TestProfileCache_Update"
	if err := vc.Update(mock.ValueDescriptorBool); err != nil {
		t.Error("failed to update value descriptor in cache")
	}

	if uv0, found := vc.ForName(mock.ValueDescriptorBool.Name); !found {
		t.Error("unable to find the value descriptor in cache after updating it")
	} else {
		assert.Equal(t, mock.ValueDescriptorBool, uv0)
	}
}
