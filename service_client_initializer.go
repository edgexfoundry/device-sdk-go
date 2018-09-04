package device

import (
	"fmt"
	"github.com/edgexfoundry/edgex-go/pkg/clients/coredata"
	logger "github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"github.com/edgexfoundry/edgex-go/pkg/clients/metadata"
	"github.com/edgexfoundry/edgex-go/pkg/clients/types"
	"net"
	"net/http"
	"strconv"
	"time"
)

// initService
// Trigger Service Client Initializer to establish connection to Metadata and Core Data Services through Metadata Client and Core Data Client.
// Service Client Initializer also needs to check the service status of Metadata and Core Data Services, because they are important dependencies of Device Service.
// The initialization process should be pending until Metadata Service and Core Data Service are both available.
func initService(useRegistry bool) {
	// check logging service
	var remoteLog = false
	var logTarget string

	if svc.c.Clients[Logging].Host == "" {
		logTarget = svc.c.Logging.File
	} else if err := checkRemoteLoggingAvailable(); err != nil {
		remoteLog = true

		host := svc.c.Clients[Logging].Host
		port := strconv.Itoa(svc.c.Clients[Logging].Port)
		addr := buildAddr(host, port)
		logTarget = addr
	} else {
		fmt.Println("Logging service is unavailable, use file instead.")
		logTarget = svc.c.Logging.File
	}

	svc.lc = logger.NewClient(svc.Name, remoteLog, logTarget)

	// check core-data core-metadata
	checkCoreDataAvailable()
	checkCoreMetaDataAvailable()

	initializeClients()
}

func checkRemoteLoggingAvailable() error {
	fmt.Println("Check Logging service is available ...")
	host := svc.c.Clients[Logging].Host
	port := strconv.Itoa(svc.c.Clients[Logging].Port)
	addr := buildAddr(host, port)
	timeout := int64(svc.c.Clients[Logging].Timeout) * int64(time.Millisecond)

	client := http.Client{
		Timeout: time.Duration(timeout),
	}

	_, err := client.Get(addr + "/api/v1" + "/ping")
	if err != nil {
		fmt.Println(fmt.Sprintf("Error getting ping: %v", err))
		return err
	}
	return nil
}

func checkCoreDataAvailable() {
	svc.lc.Info("Check CoreData service is available ...")
	host := svc.c.Clients[ClientData].Host
	port := strconv.Itoa(svc.c.Clients[ClientData].Port)
	addr := buildAddr(host, port)
	timeout := int64(svc.c.Clients[ClientData].Timeout) * int64(time.Millisecond)

	client := http.Client{
		Timeout: time.Duration(timeout),
	}

	_, err := client.Get(addr + "/api/v1" + "/ping")
	if err, ok := err.(net.Error); ok && err.Timeout() {
		svc.lc.Error(fmt.Sprintf("Timeout rror getting ping: %v", err))
		checkCoreDataAvailable()
	} else {
		svc.lc.Error(fmt.Sprintf("Error getting ping: %v", err))
		time.Sleep(10 * time.Second)
		checkCoreDataAvailable()
	}
}

func checkCoreMetaDataAvailable() {
	svc.lc.Info("Check CoreMetaData service is available ...")
	host := svc.c.Clients[ClientMetadata].Host
	port := strconv.Itoa(svc.c.Clients[ClientMetadata].Port)
	addr := buildAddr(host, port)
	timeout := int64(svc.c.Clients[ClientMetadata].Timeout) * int64(time.Millisecond)

	client := http.Client{
		Timeout: time.Duration(timeout),
	}

	_, err := client.Get(addr + "/api/v1" + "/ping")
	if err, ok := err.(net.Error); ok && err.Timeout() {
		svc.lc.Error(fmt.Sprintf("Timeout rror getting ping: %v", err))
		checkCoreDataAvailable()
	} else {
		svc.lc.Error(fmt.Sprintf("Error getting ping: %v", err))
		time.Sleep(10 * time.Second)
		checkCoreMetaDataAvailable()
	}
}

func initializeClients() {
	// initialize Core Metadata clients
	metaPort := strconv.Itoa(svc.c.Clients[ClientMetadata].Port)
	metaHost := svc.c.Clients[ClientMetadata].Host
	metaAddr := buildAddr(metaHost, metaPort)
	metaPath := v1Addressable
	metaURL := metaAddr + metaPath

	params := types.EndpointParams{
		// TODO: Can't use edgex-go internal constants!
		//ServiceKey:internal.CoreMetaDataServiceKey,
		ServiceKey:  coreMetadataServiceKey,
		Path:        metaPath,
		UseRegistry: svc.useRegistry,
		Url:         metaURL}

	svc.ac = metadata.NewAddressableClient(params, types.Endpoint{})

	params.Path = v1Device
	params.Url = metaAddr + params.Path
	svc.dc = metadata.NewDeviceClient(params, types.Endpoint{})

	params.Path = v1DevService
	params.Url = metaAddr + params.Path
	svc.sc = metadata.NewDeviceServiceClient(params, types.Endpoint{})

	params.Path = v1Deviceprofile
	params.Url = metaAddr + params.Path
	svc.dpc = metadata.NewDeviceProfileClient(params, types.Endpoint{})

	// initialize Core Data clients
	dataPort := strconv.Itoa(svc.c.Clients[ClientData].Port)
	dataHost := svc.c.Clients[ClientData].Host
	dataAddr := buildAddr(dataHost, dataPort)
	dataPath := v1Event
	dataURL := dataAddr + dataPath

	params.ServiceKey = coreDataServiceKey
	params.Path = dataPath
	params.UseRegistry = svc.useRegistry
	params.Url = dataURL

	svc.ec = coredata.NewEventClient(params, types.Endpoint{})

	params.Path = v1Valuedescriptor
	params.Url = dataAddr + dataPath
	svc.vdc = coredata.NewValueDescriptorClient(params, types.Endpoint{})
}
