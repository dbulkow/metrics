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

// Metrics demonstration. Catches messages, provides metrics.
package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var addr string

	flag.StringVar(&addr, "listen-address", ":8080", "Address to listen for HTTP requests")
	flag.Parse()

	// Create a new registry.
	reg := prometheus.NewRegistry()

	requestCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "server_http_request_count",
			Help: "The total number of requests by type",
		},
		[]string{"code", "method"},
	)

	badPathCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "server_http_request_badpath_count",
			Help: "The total number of requests to an incorrect path",
		},
		[]string{"code", "method"},
	)

	inFlightGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "server_in_flight_requests",
			Help: "A gauge of requests currently being served by the wrapped handler.",
		},
	)

	// duration is partitioned by the HTTP method and handler. It uses custom
	// buckets based on the expected request duration.
	duration := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "server_request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{0.00001, 0.0001, 0.001, 0.01, 0.1},
		},
		[]string{"handler", "method"},
	)

	// requestSize has no labels, making it a zero-dimensional
	// ObserverVec.
	requestSize := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "server_request_size_bytes",
			Help:    "A histogram of request sizes.",
			Buckets: []float64{200, 500, 900, 1500},
		},
		[]string{},
	)

	reg.MustRegister(
		collectors.NewBuildInfoCollector(),
		collectors.NewGoCollector(
			collectors.WithGoCollections(collectors.GoRuntimeMemStatsCollection|collectors.GoRuntimeMetricsCollection),
		),
		requestCount,
		inFlightGauge,
		duration,
		requestSize,
		badPathCount,
	)

	http.Handle("/submit",
		promhttp.InstrumentHandlerInFlight(inFlightGauge,
			promhttp.InstrumentHandlerCounter(requestCount,
				promhttp.InstrumentHandlerRequestSize(requestSize,
					promhttp.InstrumentHandlerDuration(duration.MustCurryWith(prometheus.Labels{"handler": "submit"}),
						http.HandlerFunc(submit),
					),
				),
			),
		),
	)

	// http.Handle("/metrics", promhttp.Handler())
	http.Handle("/metrics", promhttp.HandlerFor(
		reg,
		promhttp.HandlerOpts{
			// Opt into OpenMetrics to support exemplars.
			EnableOpenMetrics: true,
		},
	))

	http.Handle("/", promhttp.InstrumentHandlerCounter(badPathCount, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})))

	fmt.Println("listening to", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
