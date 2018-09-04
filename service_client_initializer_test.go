package device

import (
	logger "github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"testing"
)

func TestCheckCoreDataTimeout(test *testing.T) {
	var clients = map[string]service{ClientData: service{Host: "www.google.com", Port: 81, Timeout: 5000}}
	var config = Config{Clients: clients}
	svc = &Service{c: &config}
	svc.lc = logger.NewClient("test_service", false, svc.c.Logging.File)
	checkCoreDataAvailable()
}
