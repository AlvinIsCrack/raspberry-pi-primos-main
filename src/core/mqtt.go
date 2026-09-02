package core

import (
	mqtt "github.com/mochi-mqtt/server/v2"
	"github.com/mochi-mqtt/server/v2/hooks/auth"
	"github.com/mochi-mqtt/server/v2/listeners"
	"github.com/mochi-mqtt/server/v2/packets"
)

type MessageHandler func(topic string, payload []byte)

// MQTTBroker encapsula el servidor MQTT sin lógica de negocio.
type MQTTBroker struct {
	server *mqtt.Server
	addr   string
}

func NewMQTTBroker(addr string) *MQTTBroker {
	server := mqtt.New(&mqtt.Options{
		InlineClient: true,
	})
	_ = server.AddHook(new(auth.AllowHook), nil)

	return &MQTTBroker{
		server: server,
		addr:   addr,
	}
}

// OnMessage registra un handler para un patrón de tópicos.
func (b *MQTTBroker) OnMessage(topicFilter string, handler MessageHandler) {
	b.server.Subscribe(topicFilter, 1, func(cl *mqtt.Client, sub packets.Subscription, pk packets.Packet) {
		handler(pk.TopicName, pk.Payload)
	})
}

// Start arranca el listener TCP y el bucle de servicio.
func (b *MQTTBroker) Start() error {
	tcp := listeners.NewTCP(listeners.Config{
		ID:      "t1",
		Address: b.addr,
	})
	if err := b.server.AddListener(tcp); err != nil {
		return err
	}
	return b.server.Serve()
}

func (b *MQTTBroker) Stop() {
	_ = b.server.Close()
}
