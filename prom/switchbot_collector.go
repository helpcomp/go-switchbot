package prom

import (
	"context"
	"github.com/nasa9084/go-switchbot/v3/switchbot"
	"github.com/rs/zerolog/log"
)

var (
	switchbotClient       *switchbot.Client
	switchbotDevices      []switchbot.Device
	switchbotDeviceStatus map[string]switchbot.DeviceStatus
)

// New Initializes Switchbot Client
func New(token string, key string) {
	switchbotClient = switchbot.New(token, key)
	switchbotDeviceStatus = make(map[string]switchbot.DeviceStatus)
}

// Refresh Refreshes Switchbot Data
func Refresh() {
	GetDevices()
}

// GetDevices Gets all the device values
func GetDevices() {
	var err error
	var deviceStat switchbot.DeviceStatus
	switchbotDevices, _, err = switchbotClient.Device().List(context.Background())

	if err != nil {
		log.Error().Err(err).Msgf("Error Getting Devices. %s", err)
		return
	}

	// Loop through all this Switchbot Devices
	for _, d := range switchbotDevices {
		deviceStat, err = switchbotClient.Device().Status(context.Background(), d.ID)

		if err != nil {
			log.Error().Err(err).Msgf("Error Getting Device Status for %s: %s", d.Name, err)
			continue
		}
		// Get the device stats, and add them to the map
		switchbotDeviceStatus[d.ID] = deviceStat
	}
}
