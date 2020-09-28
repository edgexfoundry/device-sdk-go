package http

import (
	"net/http"

	sdkCommon "github.com/edgexfoundry/device-sdk-go/internal/common"
	"github.com/edgexfoundry/device-sdk-go/internal/container"
	"github.com/edgexfoundry/device-sdk-go/internal/telemetry"
	"github.com/edgexfoundry/go-mod-bootstrap/di"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/edgexfoundry/go-mod-core-contracts/v2/dtos/common"
)

// Ping handles the request to /ping endpoint. Is used to test if the service is working
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *V2HttpController) Ping(writer http.ResponseWriter, request *http.Request, _ *di.Container) {
	response := common.NewPingResponse()
	c.sendResponse(writer, request, contractsV2.ApiPingRoute, response, http.StatusOK)
}

// Version handles the request to /version endpoint. Is used to request the service's versions
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *V2HttpController) Version(writer http.ResponseWriter, request *http.Request, _ *di.Container) {
	response := common.NewVersionSdkResponse(sdkCommon.ServiceVersion, sdkCommon.SDKVersion)
	c.sendResponse(writer, request, contractsV2.ApiVersionRoute, response, http.StatusOK)
}

// Config handles the request to /config endpoint. Is used to request the service's configuration
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *V2HttpController) Config(writer http.ResponseWriter, request *http.Request, dic *di.Container) {
	response := common.NewConfigResponse(container.ConfigurationFrom(dic.Get))
	c.sendResponse(writer, request, contractsV2.ApiVersionRoute, response, http.StatusOK)
}

// Metrics handles the request to the /metrics endpoint, memory and cpu utilization stats
// It returns a response as specified by the V2 API swagger in openapi/v2
func (c *V2HttpController) Metrics(writer http.ResponseWriter, request *http.Request, _ *di.Container) {
	telem := telemetry.NewSystemUsage()
	metrics := common.Metrics{
		MemAlloc:       telem.Memory.Alloc,
		MemFrees:       telem.Memory.Frees,
		MemLiveObjects: telem.Memory.LiveObjects,
		MemMallocs:     telem.Memory.Mallocs,
		MemSys:         telem.Memory.Sys,
		MemTotalAlloc:  telem.Memory.TotalAlloc,
		CpuBusyAvg:     uint8(telem.CpuBusyAvg),
	}

	response := common.NewMetricsResponse(metrics)
	c.sendResponse(writer, request, contractsV2.ApiMetricsRoute, response, http.StatusOK)
}
