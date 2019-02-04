package transforms

import (
	"testing"

	"github.com/edgexfoundry/edgex-go/pkg/models"
)

func TestFilterByDeviceIDFound(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	// expectedResult := `<Event><ID></ID><Pushed>0</Pushed><Device>id1</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin><Event></Event></Event>`
	filter := Filter{
		DeviceIDs: []string{"id1"},
	}
	result := filter.FilterByDeviceID(eventIn)
	if result == nil {
		t.Fatal("result should not be nil")
	}
	if eventOut, ok := result.(*models.Event); ok {
		if eventOut.Device != "id1" {
			t.Fatal("device id does not match filter")
		}
	}
}
func TestFilterByDeviceIDNotFound(t *testing.T) {
	// Event from device 1
	eventIn := models.Event{
		Device: devID1,
	}
	// expectedResult := `<Event><ID></ID><Pushed>0</Pushed><Device>id1</Device><Created>0</Created><Modified>0</Modified><Origin>0</Origin><Event></Event></Event>`
	filter := Filter{
		DeviceIDs: []string{"id2"},
	}
	result := filter.FilterByDeviceID(eventIn)
	if result != nil {
		t.Fatal("result should be nil")
	}
}

func TestFilterByDeviceIDNoParameters(t *testing.T) {
	filter := Filter{
		DeviceIDs: []string{"id2"},
	}
	result := filter.FilterByDeviceID()
	if result != nil {
		t.Fatal("result should be nil")
	}
}
