package main

import (
	"log"
	"net/http"
	"raspberry-pi-3-sensors/internal/bme280"
	"raspberry-pi-3-sensors/internal/host"
	"raspberry-pi-3-sensors/internal/i2c"
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
	err := host.Init()
	if err != nil {
		log.Fatal(err)
	}

	i2cBus, err := i2c.NewBus(bme280I2CBus)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := i2cBus.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	bme280Device, err := bme280.NewDevice(i2cBus, bme280I2CAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		err := bme280Device.Halt()
		if err != nil {
			log.Fatal(err)
		}
	}()

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
