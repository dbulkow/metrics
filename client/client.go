// Copyright 2022 David Bulkow

// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Send a few messages to a server
package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"github.com/dbulkow/metrics"
)

func main() {
	client := &http.Client{}

	data := metrics.Data{
		ID:        45934,
		Message:   "Something Wonderful Is Happening",
		SWVersion: "35.22.1.0",
		FWVersion: "306.C1.22",
	}

	// generate an error for unsupported method
	req, err := http.NewRequest(http.MethodGet, "http://localhost:8080/submit", nil)
	if err != nil {
		log.Fatal(err)
	}

	client.Do(req)

	// Post but without Content-Type set
	buf := &bytes.Buffer{}

	err = json.NewEncoder(buf).Encode(&data)
	if err != nil {
		log.Fatal(err)
	}

	req, err = http.NewRequest(http.MethodPost, "http://localhost:8080/submit", buf)
	if err != nil {
		log.Fatal(err)
	}

	client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	// send a few thousand records
	for i := 0; i < 250000; i++ {
		buf := &bytes.Buffer{}

		err := json.NewEncoder(buf).Encode(&data)
		if err != nil {
			log.Fatal(err)
		}

		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080/submit", buf)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Add("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		if resp.StatusCode != http.StatusOK {
			log.Fatal("status ", resp.Status)
		}
	}
}
