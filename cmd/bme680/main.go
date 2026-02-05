package main

import (
	"log"
	"net/http"
	"raspberry-pi-3-sensors/internal/waterius"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.eqrx.net/mauzr/pkg/bme"
)

const i2cBus = "1"
const bme680Address = 0x77
const metricsPort = 9101
const ttl = 5

func main() {
	requests := make(chan bme.Request)
	bme.NewBME680("/dev/i2c-"+i2cBus, bme680Address, bme.Measurement{}, map[string]string{}, requests)

	collector := newPrometheusCollector(requests, ttl)
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())

	db := waterius.NewDB()
	wateriusCollector := waterius.NewPrometheusCollector(db)
	prometheus.MustRegister(wateriusCollector)

	wateriusHTTPHandler := waterius.NewHTTPHandler(db)
	http.Handle("/waterius/", wateriusHTTPHandler)

	log.Fatal(http.ListenAndServe(":"+strconv.Itoa(metricsPort), nil))
}

type prometheusCollector struct {
	temperatureMetric   *prometheus.Desc
	pressureMetric      *prometheus.Desc
	humidityMetric      *prometheus.Desc
	gasResistanceMetric *prometheus.Desc

	requests chan<- bme.Request
	ttl      int
}

func newPrometheusCollector(requests chan<- bme.Request, ttl int) *prometheusCollector {
	return &prometheusCollector{
		temperatureMetric:   prometheus.NewDesc("Temperature", "Shows temperature", nil, nil),
		pressureMetric:      prometheus.NewDesc("Pressure", "Shows pressure", nil, nil),
		humidityMetric:      prometheus.NewDesc("Humidity", "Shows humidity", nil, nil),
		gasResistanceMetric: prometheus.NewDesc("Gas", "Shows gas resistance", nil, nil),

		requests: requests,
		ttl:      ttl,
	}
}

func (c *prometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.temperatureMetric
	ch <- c.pressureMetric
	ch <- c.humidityMetric
	ch <- c.gasResistanceMetric
}

func (c *prometheusCollector) Collect(ch chan<- prometheus.Metric) {
	responses := make(chan bme.Response, 1)
	request := bme.Request{Response: responses, MaxAge: time.Now().Add(-time.Duration(c.ttl) * time.Second)}
	c.requests <- request
	response := <-responses
	if response.Err != nil {
		panic(response.Err)
	}
	measurement := response.Measurement

	temperature := prometheus.MustNewConstMetric(c.temperatureMetric, prometheus.GaugeValue, measurement.Temperature)
	temperature = prometheus.NewMetricWithTimestamp(time.Now(), temperature)

	pressure := prometheus.MustNewConstMetric(c.pressureMetric, prometheus.GaugeValue, measurement.Pressure)
	pressure = prometheus.NewMetricWithTimestamp(time.Now(), pressure)

	humidity := prometheus.MustNewConstMetric(c.humidityMetric, prometheus.GaugeValue, measurement.Humidity)
	humidity = prometheus.NewMetricWithTimestamp(time.Now(), humidity)

	gasResistance := prometheus.MustNewConstMetric(c.gasResistanceMetric, prometheus.GaugeValue, measurement.GasResistance)
	gasResistance = prometheus.NewMetricWithTimestamp(time.Now(), gasResistance)

	ch <- temperature
	ch <- pressure
	ch <- humidity
	ch <- gasResistance
}
