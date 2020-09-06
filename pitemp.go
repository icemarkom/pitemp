package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
)

const (
	defaultThermalFile = "/sys/devices/virtual/thermal/thermal_zone0/temp"
)

// Config stores the configuration for the program.
type Config struct {
	ThermalFile string
	Port        int
}

var cfg Config

func init() {
	flag.StringVar(&cfg.ThermalFile, "thermal", defaultThermalFile, "Default file containing temperature in millidegrees Celsius")
	flag.IntVar(&cfg.Port, "port", 9550, "Port to run on")
	flag.Parse()

}

// readTemperature reports temperature from cfg.ThermalFile.
func readTemperature() (int, error) {
	tf, err := ioutil.ReadFile(cfg.ThermalFile)
	if err != nil {
		return 0, fmt.Errorf("error reading thermal file %q: %w", cfg.ThermalFile, err)
	}
	t, err := strconv.Atoi(strings.TrimSpace(string(tf)))
	if err != nil {
		return 0, fmt.Errorf("error obtaining temperature from the thermal file %q: %w", cfg.ThermalFile, err)
	}
	return t, nil
}

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	t, err := readTemperature()
	if err != nil {
		log.Printf("Could not read temperature: %v.\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	fmt.Fprintln(w, t)
}

func main() {
	http.HandleFunc("/", handleRoot)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil))
}
