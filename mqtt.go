package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"sync"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

// MQTTConfig stores the MQTT-related config.
type MQTTConfig struct {
	Enabled, SSL                                                     bool
	Broker, Name, Password, Topic, TopicConfig, TopicState, Username string
	Port                                                             int
	Interval                                                         time.Duration
}

// MQTTDevice describes sensor for Home Assistant.
type MQTTDevice struct {
	Name          string `json:"name"`
	DeviceClass   string `json:"device_class"`
	Unit          string `json:"unit_of_measurement"`
	ValueTemplate string `json:"value_template"`
	StateTopic    string `json:"state_topic"`
	UniqueID      string `json:"unique_id"`
}

func handlerReconnecting(c mqtt.Client, co *mqtt.ClientOptions) {
	log.Printf("Attempting to reconnect to MQTT broker...")
}

func handlerOnConnectAttempt(b *url.URL, tc *tls.Config) *tls.Config {
	log.Println("Attempting to connect to MQTT broker...")

	return tc
}

func handlerOnConnect(c mqtt.Client) {
	log.Printf("Connection to MQTT broker established.")
}

func handlerOnConnectionLost(c mqtt.Client, e error) {
	log.Printf("Connection to MQTT broker unexpectedly lost: %v.", e)
}

func mqttClient() (mqtt.Client, error) {
	mqtt_protocol := "tcp"
	if cfg.MQTT.SSL {
		mqtt_protocol = "ssl"
	}
	urlstr := fmt.Sprintf("%s://%s:%d", mqtt_protocol, cfg.MQTT.Broker, cfg.MQTT.Port)
	mqtt_server, err := url.Parse(urlstr)
	if err != nil {
		// This is MQTT fatal.
		log.Fatalf("Error parsing server URL %q.", urlstr)
	}

	o := mqtt.NewClientOptions()
	o.Servers = append(o.Servers, mqtt_server)
	o.ClientID = cfg.MQTT.Name
	o.Username = cfg.MQTT.Username
	o.Password = cfg.MQTT.Password
	o.ConnectRetry = true
	o.AutoReconnect = true
	o.CleanSession = true
	o.ConnectRetryInterval = cfg.MQTT.Interval
	o.OnConnectAttempt = handlerOnConnectAttempt
	o.OnConnectionLost = handlerOnConnectionLost
	o.OnReconnecting = handlerReconnecting
	o.OnConnect = handlerOnConnect

	// Start the connection
	c := mqtt.NewClient(o)
	if token := c.Connect(); token.Wait() && token.Error() != nil {
		return c, fmt.Errorf("MQTT error: %v", token.Error())
	}
	return c, nil
}

func loopWait() {
	time.Sleep(cfg.MQTT.Interval)
}

func doMQTT(wg *sync.WaitGroup) {
	defer wg.Done()

	log.Printf("MQTT enabled (broker: %q, port %d, SSL: %v, username: %q, topic prefix: %q.",
		cfg.MQTT.Broker, cfg.MQTT.Port, cfg.MQTT.SSL, cfg.MQTT.Username, cfg.MQTT.Topic)

	cfg.MQTT.TopicConfig = fmt.Sprintf("%s/config", cfg.MQTT.Topic)
	cfg.MQTT.TopicState = fmt.Sprintf("%s/state", cfg.MQTT.Topic)

	d := MQTTDevice{
		Name:          fmt.Sprintf("%s_temperature", cfg.MQTT.Name),
		DeviceClass:   "temperature",
		Unit:          fmt.Sprintf("%s%s", cfg.UnitPrefix, cfg.Unit),
		ValueTemplate: "{{ value_json.temperature | round(1) }}",
		StateTopic:    cfg.MQTT.TopicState,
		UniqueID:      fmt.Sprintf("%s_temperature", cfg.MQTT.Name),
	}

	dr, err := json.Marshal(d)
	if err != nil {
		log.Fatalf("Could not generate JSON response: %v", err)
	}

	c, err := mqttClient()
	if err != nil {
		log.Fatalf("Could not create MQTT client: %v", err)
	}
	for {
		if !c.IsConnected() || !c.IsConnectionOpen() {
			loopWait()
			continue
		}
		jr, t, err := JSONResponse(cfg.MQTT.Name)
		if err != nil {
			log.Printf("Could not read temperature: %v.\n", err)
			loopWait()
			continue
		}
		if token := c.Publish(cfg.MQTT.TopicConfig, 0, false, dr); token.Wait() && token.Error() != nil {
			log.Printf("MQTT publish error: %v", token.Error())
			loopWait()
			continue
		}
		if token := c.Publish(cfg.MQTT.TopicState, 0, false, jr); token.Wait() && token.Error() != nil {
			log.Printf("MQTT publish error: %v", token.Error())
		}
		log.Printf("Reported to MQTT broker %q on %q, temperature %.3f %s%s.", cfg.MQTT.Broker, cfg.MQTT.TopicState, t, cfg.UnitPrefix, cfg.Unit)
		loopWait()
	}
}
