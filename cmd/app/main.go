package main

import (
	"fmt"
	"log"
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
	"periph.io/x/conn/v3/physic"
	"periph.io/x/devices/v3/bmxx80"
	"periph.io/x/host/v3"
)

const i2cBus = "1"
const bme280I2cAddress = 0x76

func main() {
	err := initHost()
	if err != nil {
		log.Fatal(err)
	}

	bus, err := newBus(i2cBus)
	if err != nil {
		log.Fatal(err)
	}
	defer bus.Close()

	device, err := newDevice(bus, bme280I2cAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer device.Halt()

	env, err := newEnv(device)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%8s %10s %9s\n", env.Temperature, env.Pressure, env.Humidity)
}

func initHost() error {
	// Load all the drivers:
	if _, err := host.Init(); err != nil {
		return err
	}

	return nil
}

func newBus(i2cBus string) (i2c.BusCloser, error) {
	// Open a handle to the first available I²C bus:
	bus, err := i2creg.Open(i2cBus)
	if err != nil {
		return nil, err
	}

	return bus, err
}

func newDevice(bus i2c.Bus, address uint16) (*bmxx80.Dev, error) {
	// Open a handle to a bme280/bmp280 connected on the²C bus using default
	// settings:
	device, err := bmxx80.NewI2C(bus, address, &bmxx80.DefaultOpts)
	if err != nil {
		return nil, err
	}

	return device, err
}

func newEnv(device *bmxx80.Dev) (*physic.Env, error) {
	// Read temperature from the sensor:
	var env physic.Env
	if err := device.Sense(&env); err != nil {
		return nil, err
	}

	return &env, nil
}
