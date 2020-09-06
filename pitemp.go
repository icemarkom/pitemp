package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	defaultThermalFile = "/sys/devices/virtual/thermal/thermal_zone0/temp"
)

// Config stores the configuration for the program.
type Config struct {
	ThermalFile string
	Port        int
}

// JSONReturn holds the values for JSON printout.
type JSONReturn struct {
	Temperature int    `json:"temperature"`
	Requestor   string `json:"requestor"`
	ServerTime  string `json:"time"`
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
	var (
		j   JSONReturn
		err error
	)

	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	j.Temperature, err = readTemperature()
	if err != nil {
		log.Printf("Could not read temperature: %v.\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	j.Requestor = r.RemoteAddr
	j.ServerTime = time.Now().Format("2006-01-02 15:04:05")
	jr, err := json.Marshal(j)
	if err != nil {
		log.Printf("Could not generate JSON response: %v", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, string(jr))
	log.Printf("request from: %s, reported temperature %d", r.RemoteAddr, j.Temperature)
}

func main() {
	http.HandleFunc("/", handleRoot)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), nil))
}
