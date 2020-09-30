package http

import (
	"net/http"

	sdkCommon "github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/handler"
	edgexErr "github.com/edgexfoundry/go-mod-core-contracts/errors"
)

func (c *V2HttpController) Discovery(writer http.ResponseWriter, request *http.Request) {
	ds := container.DeviceServiceFrom(c.dic.Get)
	correlationID := request.Header.Get(sdkCommon.CorrelationHeader)

	err := checkServiceLocked(request, ds.AdminState)
	if err != nil {
		c.sendError(writer, request, edgexErr.KindServiceLocked, "Service Locked", err, sdkCommon.APIV2DiscoveryRoute, correlationID)
		return
	}

	configuration := container.ConfigurationFrom(c.dic.Get)
	if !configuration.Device.Discovery.Enabled {
		c.sendError(writer, request, edgexErr.KindServiceUnavailable, "Device discovery disabled", nil, sdkCommon.APIV2DiscoveryRoute, correlationID)
		return
	}

	discovery := container.ProtocolDiscoveryFrom(c.dic.Get)
	if discovery == nil {
		c.sendError(writer, request, edgexErr.KindNotImplemented, "ProtocolDiscovery not implemented", nil, sdkCommon.APIV2DiscoveryRoute, correlationID)
		return
	}

	handler.DiscoveryHandler(writer, discovery, c.lc)
}
