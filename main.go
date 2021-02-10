package main

import (
	"github.com/lmarszal/tcpstat_k8s_exporter/collector"
	"github.com/lmarszal/tcpstat_k8s_exporter/docker"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log"
	"net/http"
	"os"
)

func main() {
	dockercli, err := docker.New()
	if err != nil {
		panic(err)
	}
	tcpstatscollector := collector.New(&dockercli, os.Getenv("NAMESPACE"))
	prometheus.MustRegister(tcpstatscollector)

	http.Handle("/metrics", promhttp.Handler())
	log.Println("Listening on port 8080...")
	_ = http.ListenAndServe(":8080", nil)
}
