package prometheus

import (
	"context"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/bootstrap/startup"
	"github.com/edgexfoundry/go-mod-bootstrap/v2/di"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"sync"
)

func BootstrapHandler(
	ctx context.Context,
	wg *sync.WaitGroup,
	_ startup.Timer,
	dic *di.Container) bool {

	go func (){
		r := mux.NewRouter()
		r.Handle("/metrics",promhttp.Handler())
		http.ListenAndServe(":9090",r)
	}()

	return true
}
