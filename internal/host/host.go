package host

import "periph.io/x/host/v3"

func Init() error {
	if _, err := host.Init(); err != nil {
		return err
	}

	return nil
}
