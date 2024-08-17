package main

import (
	"flag"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"math/rand"
	"net/http"
	"time"
)

type demoAPI struct {
	registry         *prometheus.Registry
	requestDurations *prometheus.SummaryVec
	jobCounter       *prometheus.CounterVec
}

func (a demoAPI) register(mux *http.ServeMux) {
	// HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
	mux.HandleFunc("/api/foo", a.foo)
	mux.HandleFunc("/api/bar", a.bar)
	// Handle(pattern string, handler http.Handler)
	mux.Handle("/metrics", promhttp.HandlerFor(a.registry, promhttp.HandlerOpts{}))
}

func (a demoAPI) foo(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(a.requestDurations.WithLabelValues("/foo"))
	log.Println("Handling foo...")

	// Simulate a random duration that the "foo" operation needs to be completed.
	time.Sleep(25*time.Millisecond + time.Duration(rand.Float64()*150)*time.Millisecond)

	w.Write([]byte("Handled foo"))
	timer.ObserveDuration()
}

func (a demoAPI) bar(w http.ResponseWriter, r *http.Request) {
	timer := prometheus.NewTimer(a.requestDurations.WithLabelValues("/bar"))
	log.Println("Handling bar...")
	// Simulate a random duration that the "bar" operation needs to be completed.
	time.Sleep(50*time.Millisecond + time.Duration(rand.Float64()*200)*time.Millisecond)

	w.Write([]byte("Handled bar"))
	timer.ObserveDuration()
}

func periodicBackgroundTask(jobCounter *prometheus.CounterVec) {
	log.Println("Starting background task loop...")
	bgTicker := time.NewTicker(5 * time.Second)
	for {
		jobCounter.WithLabelValues("total").Inc()
		log.Println("Performing background task...")
		// Simulate a random duration that the background task needs to be completed.
		time.Sleep(1*time.Second + time.Duration(rand.Float64()*500)*time.Millisecond)

		// Simulate the background task either succeeding or failing (with a 30% probability).
		if rand.Float64() > 0.3 {
			log.Println("Background task completed successfully.")
		} else {
			log.Println("Background task failed.")
			jobCounter.WithLabelValues("failed").Inc()
		}

		<-bgTicker.C
	}
}

func main() {
	registry := prometheus.NewRegistry()

	requestDurations := prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name: "http_request_duration_seconds",
		Help: "A summary of the HTTP request duration in seconds.",
		Objectives: map[float64]float64{
			0.5:  0.05,
			0.9:  0.01,
			0.99: 0.001},
	},
		[]string{"path"},
	)

	jobCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "batchjob_count",
		Help: "A counter of batch job runs, success and failed",
	},
		[]string{"status"},
	)

	registry.MustRegister(requestDurations, jobCounter)

	listenAddr := flag.String("web.listen-addr", ":8080", "The address to listen on for web requests.")
	flag.Parse()

	go periodicBackgroundTask(jobCounter)

	api := &demoAPI{requestDurations: requestDurations, registry: registry, jobCounter: jobCounter}
	api.register(http.DefaultServeMux)

	log.Fatal(http.ListenAndServe(*listenAddr, nil))
}
