package bme280

import (
	"fmt"
	"log"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/bmxx80"
	"periph.io/x/host/v3"
)

func NewDevice(i2cBus string, i2cAddress uint16) (*bmxx80.Dev, error) {
	if _, err := host.Init(); err != nil {
		return nil, err
	}

	bus, err := i2creg.Open(i2cBus)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := bus.Close()
		if err != nil {
			panic(err)
		}
	}()

	device, err := bmxx80.NewI2C(bus, i2cAddress, &bmxx80.DefaultOpts)
	if err != nil {
		return nil, err
	}
	defer func() {
		err := device.Halt()
		if err != nil {
			panic(err)
		}
	}()

	return device, nil
}

type PrometheusCollector struct {
	temperatureMetric *prometheus.Desc
	pressureMetric    *prometheus.Desc
	humidityMetric    *prometheus.Desc

	device *bmxx80.Dev
}

func NewPrometheusCollector(device *bmxx80.Dev, metricsPrefix string) *PrometheusCollector {
	return &PrometheusCollector{
		temperatureMetric: prometheus.NewDesc(
			fmt.Sprintf("%s_temperature", metricsPrefix), "Shows temperature", nil, nil,
		),
		pressureMetric: prometheus.NewDesc(
			fmt.Sprintf("%s_pressure", metricsPrefix), "Shows pressure", nil, nil,
		),
		humidityMetric: prometheus.NewDesc(
			fmt.Sprintf("%s_humidity", metricsPrefix), "Shows humidity", nil, nil,
		),

		device: device,
	}
}

func (c *PrometheusCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.temperatureMetric
	ch <- c.pressureMetric
	ch <- c.humidityMetric
}

func (c *PrometheusCollector) Collect(ch chan<- prometheus.Metric) {
	var env physic.Env
	if err := c.device.Sense(&env); err != nil {
		log.Fatal(err)
	}

	temperature := prometheus.NewMetricWithTimestamp(
		time.Now(),
		prometheus.MustNewConstMetric(c.temperatureMetric, prometheus.GaugeValue, float64(env.Temperature)),
	)

	pressure := prometheus.NewMetricWithTimestamp(
		time.Now(),
		prometheus.MustNewConstMetric(c.pressureMetric, prometheus.GaugeValue, float64(env.Pressure)),
	)

	humidity := prometheus.NewMetricWithTimestamp(
		time.Now(),
		prometheus.MustNewConstMetric(c.humidityMetric, prometheus.GaugeValue, float64(env.Humidity)),
	)

	ch <- temperature
	ch <- pressure
	ch <- humidity
}
