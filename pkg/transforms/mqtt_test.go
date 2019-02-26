package transforms

import (
	"strconv"
	"strings"
	"testing"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/edgexfoundry/app-functions-sdk-go/pkg/excontext"
	logger "github.com/edgexfoundry/go-mod-core-contracts/clients/logging"
	"github.com/edgexfoundry/go-mod-core-contracts/models"
	"github.com/stretchr/testify/assert"
)

var addr models.Addressable

func init() {
	lc = logger.NewClient("app_functions_sdk_go", false, "./test.log", "DEBUG")
	addr = models.Addressable{
		Address:   "localhost",
		Port:      1883,
		Protocol:  "tcp",
		Publisher: "",
		Password:  "",
		Topic:     "testMQTTTopic",
	}
}

func TestMQTTSend(t *testing.T) {
	t.SkipNow()
	protocol := strings.ToLower(addr.Protocol)
	opts := MQTT.NewClientOptions()
	broker := protocol + "://" + addr.Address + ":" + strconv.Itoa(addr.Port) + addr.Path
	opts.AddBroker(broker)
	opts.SetClientID(addr.Publisher)
	opts.SetUsername(addr.User)
	opts.SetPassword(addr.Password)
	opts.SetAutoReconnect(false)

	sender := MQTTSender{
		client: MQTT.NewClient(opts),
		topic:  "",
	}
	ctx := excontext.Context{
		LoggingClient: lc,
	}
	dataToSend := "SOME DATA TO SEND"
	continuePipeline, result := sender.MQTTSend(ctx, dataToSend)
	assert.Equal(t, true, continuePipeline, "Should Continue Pipeline")
	assert.Equal(t, nil, result, "Should be nil")

}
func TestMQTTSendNoData(t *testing.T) {

	sender := MQTTSender{}
	ctx := excontext.Context{
		LoggingClient: lc,
	}
	continuePipeline, result := sender.MQTTSend(ctx)
	assert.Equal(t, false, continuePipeline, "Should Not Continue Pipeline")
	assert.Equal(t, "No Data Received", result.(error).Error(), "Error should be: No Data Received")

}
func TestMQTTSendInvalidData(t *testing.T) {
	t.SkipNow()

	protocol := strings.ToLower(addr.Protocol)
	opts := MQTT.NewClientOptions()
	broker := protocol + "://" + addr.Address + ":" + strconv.Itoa(addr.Port) + addr.Path
	opts.AddBroker(broker)
	opts.SetClientID(addr.Publisher)
	opts.SetUsername(addr.User)
	opts.SetPassword(addr.Password)
	opts.SetAutoReconnect(false)

	sender := MQTTSender{
		client: MQTT.NewClient(opts),
		topic:  "",
	}
	ctx := excontext.Context{
		LoggingClient: lc,
	}
	dataToSend := "SOME DATA TO SEND"
	continuePipeline, result := sender.MQTTSend(ctx, ([]byte)(dataToSend))
	assert.Equal(t, false, continuePipeline, "Should Not Continue Pipeline")
	assert.Equal(t, "Unexpected type received", result.(error).Error(), "Error should be: Unexpected type received")

}

func TestNewMQTTSender(t *testing.T) {
	addr1 := models.Addressable{
		Address:   "localhost",
		Port:      1883,
		Protocol:  "tcp",
		Path:      "/path",
		Publisher: "publisher",
		User:      "user",
		Password:  "password",
		Topic:     "testMQTTTopic",
	}
	sender := NewMQTTSender(lc, addr1, "", "")
	assert.NotNil(t, sender.client, "Client should not be nil")
	opts := sender.client.OptionsReader()
	assert.Equal(t, "testMQTTTopic", sender.topic, "Topic should be set to testMQTTTopic")
	servers := opts.Servers()
	assert.Equal(t, 1, len(servers), "Should have 1 server connection defined")
	assert.Equal(t, "localhost:1883", servers[0].Host, "Server connection to host should be: localhost:1883")
	assert.Equal(t, "tcp", servers[0].Scheme, "Server connection protocol should be tcp")
	assert.Equal(t, "/path", servers[0].Path, "Server connection path should be path")
	assert.Equal(t, "publisher", opts.ClientID(), "ClientID should be publisher")
	assert.Equal(t, "user", opts.Username(), "Username should be user")
	assert.Equal(t, "password", opts.Password(), "Password should be password")
	assert.Equal(t, false, opts.AutoReconnect(), "Autoreconnect should be false")
}
