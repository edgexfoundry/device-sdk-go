//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package v2

import (
	"net/http"

	v2http "github.com/edgexfoundry/app-functions-sdk-go/internal/v2/controller/http"
	"github.com/edgexfoundry/go-mod-core-contracts/clients/logger"
	contractsV2 "github.com/edgexfoundry/go-mod-core-contracts/v2"
	"github.com/gorilla/mux"
)

// ConfigureStandardRoutes loads standard V2 routes
func ConfigureStandardRoutes(router *mux.Router, lc logger.LoggingClient) {
	controller := v2http.NewV2Controller(lc)

	lc.Info("Registering standard V2 routes...")

	router.HandleFunc(contractsV2.ApiPingRoute, controller.Ping).Methods(http.MethodGet)
	router.HandleFunc(contractsV2.ApiVersionRoute, controller.Version).Methods(http.MethodGet)
	router.HandleFunc(contractsV2.ApiMetricsRoute, controller.Metrics).Methods(http.MethodGet)
}
