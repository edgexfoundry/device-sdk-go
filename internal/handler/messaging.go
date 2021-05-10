package handler

import (
	"context"
	"sync"

	"github.com/edgexfoundry/device-sdk-go/v2/internal/container"
	bootstrapContainer "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/container"
	bootstrapMessaging "github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/messaging"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
)

func MessagingBootstrapHandler(ctx context.Context, wg *sync.WaitGroup, startupTimer startup.Timer, dic *di.Container) bool {
	lc := bootstrapContainer.LoggingClientFrom(dic.Get)
	config := container.ConfigurationFrom(dic.Get)
	if config.Service.UseMessageBus == true {
		return bootstrapMessaging.BootstrapHandler(ctx, wg, startupTimer, dic)
	}

	lc.Info("Use of MessageBus disabled, skipping creation of messaging client")
	return true
}
