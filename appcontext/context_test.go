package appcontext

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/edgexfoundry/go-mod-core-contracts/clients"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/coredata"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/types"
	"github.com/stretchr/testify/assert"
)

func TestComplete(t *testing.T) {
	ctx := Context{}
	testData := "output data"
	ctx.Complete([]byte(testData))
	assert.Equal(t, []byte(testData), ctx.OutputData)
}

var eventClient coredata.EventClient
var params types.EndpointParams

func init() {
	params = types.EndpointParams{
		ServiceKey:  clients.CoreDataServiceKey,
		Path:        clients.ApiEventRoute,
		UseRegistry: false,
		Url:         "http://test" + clients.ApiEventRoute,
		Interval:    clients.ClientMonitorDefault,
	}

}
func TestMarkAsPushedNoEventIdOrChecksum(t *testing.T) {
	ctx := Context{}
	err := ctx.MarkAsPushed()
	assert.NotNil(t, err)
	assert.Equal(t, "No EventID or EventChecksum Provided", err.Error())
}

func TestMarkAsPushedNoChecksum(t *testing.T) {
	testChecksum := "checksumValue"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method != http.MethodPut {
			t.Errorf("expected http method is PUT, active http method is : %s", r.Method)
		}
		url := clients.ApiEventRoute + "/checksum/" + testChecksum
		if r.URL.EscapedPath() != url {
			t.Errorf("expected uri path is %s, actual uri path is %s", url, r.URL.EscapedPath())
		}
	}))

	defer ts.Close()
	params.Url = ts.URL + clients.ApiEventRoute
	eventClient = coredata.NewEventClient(params, mockEventEndpoint{})
	ctx := Context{
		EventChecksum: testChecksum,
		CorrelationID: "correlationId",
		EventClient:   eventClient,
	}
	err := ctx.MarkAsPushed()

	assert.Nil(t, err)

}

func TestMarkAsPushedEventId(t *testing.T) {
	testID := "eventId"

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)

		if r.Method != http.MethodPut {
			t.Errorf("expected http method is PUT, active http method is : %s", r.Method)
		}
		//"http://test/api/v1/event/id/eventId"
		url := clients.ApiEventRoute + "/id/" + testID
		if r.URL.EscapedPath() != url {
			t.Errorf("expected uri path is %s, actual uri path is %s", url, r.URL.EscapedPath())
		}
	}))

	defer ts.Close()
	params.Url = ts.URL + clients.ApiEventRoute
	eventClient = coredata.NewEventClient(params, mockEventEndpoint{})

	ctx := Context{
		EventID:       testID,
		CorrelationID: "correlationId",
		EventClient:   eventClient,
	}

	err := ctx.MarkAsPushed()

	assert.Nil(t, err)
}

type mockEventEndpoint struct {
}

func (e mockEventEndpoint) Monitor(params types.EndpointParams, ch chan string) {
	switch params.ServiceKey {
	case clients.CoreDataServiceKey:
		url := fmt.Sprintf("http://%s:%v%s", "localhost", 48080, params.Path)
		ch <- url
		break
	default:
		ch <- ""
	}
}
