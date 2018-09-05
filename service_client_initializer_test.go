package device

import (
	"github.com/edgexfoundry/edgex-go/pkg/clients/logging"
	"net"
	"testing"
)

func TestCheckServiceUpByPingWithTimeoutError(test *testing.T) {
	var clients = map[string]service{ClientData: service{Host: "www.google.com", Port: 81, Timeout: 5000}}
	var config = Config{Clients: clients}
	svc = &Service{c: &config}
	svc.lc = logger.NewClient("test_service", false, svc.c.Logging.File)

	err := checkServiceUpByPing(ClientData)

	if err, ok := err.(net.Error); ok && !err.Timeout() {
		test.Fatal("Should be timeout error")
	}

}
