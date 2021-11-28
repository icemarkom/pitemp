// Copyright 2021 PiTemp Contributors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	defaultThermalFile  = `/sys/devices/virtual/thermal/thermal_zone0/temp`
	defaultMQTTTopic    = `homeassistant/sensor/%s_temperature`
	defaultHTTPPort     = 9550
	defaultMQTTPort     = 1883
	defaultMQTTSSLPort  = 8883
	defaultMQTTInterval = 30
	mqtt_client_regex   = `^[A-Za-z0-9_-]+$`
)

var version, gitCommit string

// HTTPConfig stores the HTTP-related config.
type HTTPConfig struct {
	Enabled bool
	Port    int
}

// Config stores the configuration for the program.
type Config struct {
	ThermalFile string
	Unit        string
	UnitPrefix  string
	HTTP        HTTPConfig
	MQTT        MQTTConfig
}

// JSONReturn holds the values for JSON printout.
type JSONReturn struct {
	Temperature float64 `json:"temperature"`
	Requestor   string  `json:"requestor"`
	ServerTime  string  `json:"time"`
}

var cfg Config

func validateFlags() error {
	switch cfg.Unit {
	case "c", "C":
		cfg.UnitPrefix = "°"
		cfg.Unit = "C"
	case "f", "F":
		cfg.UnitPrefix = "°"
		cfg.Unit = "F"
	case "k", "K":
		cfg.Unit = "K"
	default:
		return fmt.Errorf("invalid temperature unit %q", cfg.Unit)
	}

	// MQTT related checks.
	if cfg.MQTT.Enabled {
		// Broken must be specified.
		if cfg.MQTT.Broker == "" {
			return errors.New("MQTT broker must be specified")
		}
		// Port must be non-zero.
		if cfg.MQTT.Port == 0 {
			return errors.New("MQTT broken port must not be 0")
		}
		// Client name must be set.
		if cfg.MQTT.Name == "" {
			return errors.New("MQTT client name must be specified")
		}
		// Client name must only use specific characters.
		cre := regexp.MustCompile(mqtt_client_regex)
		if !cre.MatchString(cfg.MQTT.Name) {
			return fmt.Errorf("%q is an invalid MQTT client name; it should match %q", cfg.MQTT.Name, mqtt_client_regex)
		}
	}
	return nil
}

func isFlagPassed(fl string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == fl {
			found = true
		}
	})
	return found
}

func printVersion(v, g string) {
	fmt.Printf("Version: %s\n Commit: %s\n", v, g)
}

func init() {
	var v bool

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Cannot obtain hostname: %v.", err)
	}
	hostname = strings.Split(hostname, ".")[0]

	// For now, HTTP is always enabled.
	cfg.HTTP.Enabled = true

	flag.StringVar(&cfg.ThermalFile, "thermal_file", defaultThermalFile, "Default file containing temperature in millidegrees Celsius")
	flag.StringVar(&cfg.Unit, "unit", "C", "C or F")
	flag.IntVar(&cfg.HTTP.Port, "http_port", defaultHTTPPort, "HTTP port to listen on")
	flag.BoolVar(&cfg.MQTT.Enabled, "mqtt", false, "Notify MQTT broker")
	flag.StringVar(&cfg.MQTT.Broker, "mqtt_broker", "localhost", "MQTT Broker address")
	flag.IntVar(&cfg.MQTT.Port, "mqtt_port", defaultMQTTPort, "MQTT Broker port (will use 8883 if SSL is enabled)")
	flag.BoolVar(&cfg.MQTT.SSL, "mqtt_ssl", false, "MQTT SSL")
	flag.StringVar(&cfg.MQTT.Name, "mqtt_client", hostname, "MQTT Client Name")
	flag.StringVar(&cfg.MQTT.Topic, "mqtt_topic", fmt.Sprintf(defaultMQTTTopic, hostname), "MQTT Topic prefix")
	flag.DurationVar(&cfg.MQTT.Interval, "mqtt_interval", defaultMQTTInterval*time.Second, "MQTT notification interval in seconds")
	flag.StringVar(&cfg.MQTT.Username, "mqtt_username", "", "MQTT username")
	flag.StringVar(&cfg.MQTT.Password, "mqtt_password", "", "MQTT password")
	flag.BoolVar(&v, "version", false, "Print version")
	flag.Parse()

	if v {
		printVersion(version, gitCommit)
		os.Exit(42)
	}

	if !isFlagPassed("mqtt_port") {
		if cfg.MQTT.SSL {
			cfg.MQTT.Port = defaultMQTTSSLPort
		}
	}

	if cfg.MQTT.Name != hostname {
		if cfg.MQTT.Topic == fmt.Sprintf(defaultMQTTTopic, hostname) {
			cfg.MQTT.Topic = fmt.Sprintf(defaultMQTTTopic, cfg.MQTT.Name)
		}
	}

	if err = validateFlags(); err != nil {
		log.Fatalf("Incorrect use of flags: %v.", err)
	}
}

func Fahrenheit(t float64) float64 {
	return (t * 9.0 / 5.0) + 32.0
}

func Kelvin(t float64) float64 {
	return t + 273.0
}

// readTemperature reports temperature from cfg.ThermalFile.
func readTemperature() (float64, error) {
	tf, err := ioutil.ReadFile(cfg.ThermalFile)
	if err != nil {
		return 0, fmt.Errorf("error reading thermal file %q: %w", cfg.ThermalFile, err)
	}
	t, err := strconv.Atoi(strings.TrimSpace(string(tf)))
	if err != nil {
		return 0, fmt.Errorf("error obtaining temperature from the thermal file %q: %w", cfg.ThermalFile, err)
	}
	// Temperature is stored in millidegrees Celsius. We are OK with a rounded number.
	rt := float64(t) / 1000.0
	switch cfg.Unit {
	case "F":
		return Fahrenheit(rt), nil
	case "K":
		return Kelvin(rt), nil
	default:
		return rt, nil
	}
}

func JSONResponse(requestor string) (string, float64, error) {
	var (
		j   JSONReturn
		err error
	)

	j.Requestor = requestor
	j.Temperature, err = readTemperature()
	if err != nil {
		return "", 0, fmt.Errorf("could not read temperature: %v", err)
	}
	j.ServerTime = time.Now().Format("2006-01-02 15:04:05")
	jr, err := json.Marshal(j)
	if err != nil {
		return "", 0, fmt.Errorf("could not generate JSON response: %v", err)
	}
	return string(jr), j.Temperature, nil
}

func main() {
	wg := new(sync.WaitGroup)

	log.Printf("Thermal file: %q.", cfg.ThermalFile)
	// Run our HTTP stuff.
	if cfg.HTTP.Enabled {
		wg.Add(1)
		go doHTTP(wg)
	}

	// Should we do MQTT?
	if cfg.MQTT.Enabled {
		wg.Add(1)
		go doMQTT(wg)
	}
	wg.Wait()
}
