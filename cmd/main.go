package main

import (
	"log"
	"net/http"
	"raspberry-pi-3-sensors/internal/bme280"
	"raspberry-pi-3-sensors/internal/waterius"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const bme280I2CBus = "1"
const bme280I2CAddress = 0x76
const metricsPrefix = "sensors_0"
const metricsPort = 9101

func main() {
	bme280Device, err := bme280.NewDevice(bme280I2CBus, bme280I2CAddress)
	if err != nil {
		log.Fatal(err)
	}

	bme280Collector := bme280.NewPrometheusCollector(bme280Device, metricsPrefix)
	prometheus.MustRegister(bme280Collector)

	db := waterius.NewDB()
	wateriusCollector := waterius.NewPrometheusCollector(db)
	prometheus.MustRegister(wateriusCollector)
	wateriusHTTPHandler := waterius.NewHTTPHandler(db)

	http.Handle("/waterius/", wateriusHTTPHandler)
	http.Handle("/metrics", promhttp.Handler())

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(metricsPort), nil))
}
