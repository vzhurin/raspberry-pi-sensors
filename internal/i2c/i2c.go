package i2c

import (
	"periph.io/x/conn/v3/i2c"
	"periph.io/x/conn/v3/i2c/i2creg"
)

func NewBus(i2cBus string) (i2c.BusCloser, error) {
	bus, err := i2creg.Open(i2cBus)
	if err != nil {
		return nil, err
	}

	return bus, nil
}
