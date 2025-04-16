package gofactory

import (
	"time"

	"github.com/xyzj/mqtt-server"
)

func (s *Service) MqttBrokerWrite(topic string, body []byte, qos byte) error {
	return s.mqttbroker.Publish(topic, body, qos)
}

func (s *Service) MqttBrokerSubscribe(topic string, subscriptionId int, handler mqtt.InlineSubFn) error {
	return s.mqttbroker.Subscribe(topic, subscriptionId, handler)
}

func (s *Service) MqttWrite(topic string, body []byte, qos byte) error {
	return s.opt.climqtt.cli.WriteWithQos(topic, body, qos)
}

func (s *Service) RMQWrite(topic string, body []byte, expire time.Duration) {
	s.opt.clirmq.clip.Send(topic, body, expire)
}
