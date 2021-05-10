package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"
)

func handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	jr, t, err := JSONResponse(r.RemoteAddr)
	if err != nil {
		log.Printf("Could not read temperature: %v.\n", err)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}
	fmt.Fprintln(w, jr)
	log.Printf("Request from: %s, reported temperature %.3f %s%s.", r.RemoteAddr, t, cfg.UnitPrefix, cfg.Unit)
}

func doHTTP(wg *sync.WaitGroup) {
	http.HandleFunc("/", handleRoot)
	log.Printf("HTTP enabled (port: %d).", cfg.HTTP.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.HTTP.Port), nil))
	wg.Done()
}
