package transforms

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/diegoholiveira/jsonlogic"
	"github.com/edgexfoundry/app-functions-sdk-go/appcontext"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/util"
)

// JSONLogic ...
type JSONLogic struct {
	Rule *strings.Reader
}

// NewJSONLogic creates, initializes and returns a new instance of HTTPSender
func NewJSONLogic(rule string) JSONLogic {
	return JSONLogic{
		Rule: strings.NewReader(rule),
	}
}

// Evaluate ...
func (logic JSONLogic) Evaluate(edgexcontext *appcontext.Context, params ...interface{}) (bool, interface{}) {
	if len(params) < 1 {
		// We didn't receive a result
		return false, errors.New("No Data Received")
	}

	coercedData, err := util.CoerceType(params[0])
	if err != nil {
		return false, err
	}

	data := strings.NewReader(string(coercedData))
	var logicresult bytes.Buffer
	edgexcontext.LoggingClient.Debug("Applying JSONLogic Rule")
	err = jsonlogic.Apply(logic.Rule, data, &logicresult)
	if err != nil {
		return false, err
	}
	var result bool
	decoder := json.NewDecoder(&logicresult)
	decoder.Decode(&result)
	edgexcontext.LoggingClient.Debug("Condition met: %" + strconv.FormatBool(result))

	return result, params[0]
}
