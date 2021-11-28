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
