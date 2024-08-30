package prom

import (
	"fmt"
	"github.com/nasa9084/go-switchbot/v3/switchbot"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "switchbot"
)

type Exporter struct {
	DeviceInfo         *prometheus.Desc
	DeviceCalibrated   *prometheus.Desc
	DeviceCloudEnabled *prometheus.Desc
	DeviceBattery      *prometheus.Desc
	LockLockState      *prometheus.Desc
	LockDoorState      *prometheus.Desc
}

func (e *Exporter) Collect(metrics chan<- prometheus.Metric) {
	Refresh() // Refreshes Data

	// Loop through all the devices
	for _, device := range switchbotDevices {
		// Device Description
		metrics <- prometheus.MustNewConstMetric(
			e.DeviceInfo,
			prometheus.GaugeValue,
			1,
			device.ID, device.Name, string(device.Type), device.GroupName, device.Hub, string(switchbotDeviceStatus[device.ID].Version), // {"id", "name", "type", "group", "hub_id", "version"}
		)
		metrics <- prometheus.MustNewConstMetric(
			e.DeviceCloudEnabled,
			prometheus.GaugeValue,
			Bool2f64(device.IsEnableCloudService),
			device.ID, device.Name,
		)
		metrics <- prometheus.MustNewConstMetric(
			e.DeviceCalibrated,
			prometheus.GaugeValue,
			Bool2f64(switchbotDeviceStatus[device.ID].IsCalibrated),
			device.ID, device.Name,
		)

		// Device-Specific Metrics
		switch device.Type {
		//#region SmartLockPro
		case switchbot.SmartLockPro:
			metrics <- prometheus.MustNewConstMetric(
				e.DeviceBattery,
				prometheus.GaugeValue,
				float64(switchbotDeviceStatus[device.ID].Battery),
				device.ID, device.Name,
			)
			metrics <- prometheus.MustNewConstMetric(
				e.LockLockState,
				prometheus.GaugeValue,
				StateOK(switchbotDeviceStatus[device.ID].LockState),
				device.ID, device.Name, switchbotDeviceStatus[device.ID].LockState,
			)
			metrics <- prometheus.MustNewConstMetric(
				e.LockDoorState,
				prometheus.GaugeValue,
				StateOK(switchbotDeviceStatus[device.ID].DoorState),
				device.ID, device.Name, switchbotDeviceStatus[device.ID].DoorState,
			)
			//#endregion
		}
	}
}

// Describe Prometheus Describer
func (e *Exporter) Describe(descs chan<- *prometheus.Desc) {
	descs <- e.DeviceInfo
	descs <- e.DeviceCalibrated
	descs <- e.DeviceCloudEnabled
	descs <- e.DeviceBattery
	descs <- e.LockLockState
	descs <- e.LockDoorState
}

// NewExporter New Prometheus Exporter
func NewExporter() *Exporter {
	return &Exporter{
		DeviceInfo: prometheus.NewDesc(
			prometheus.BuildFQName(
				namespace,
				"info",
				"description",
			),
			"Information of the gathered devices",
			[]string{"id", "name", "type", "group", "hub_id", "version"},
			nil,
		),
		DeviceCalibrated:   prometheusDevice("calibrated", "determines if the open position and the close position of a device have been properly calibrated or not"),
		DeviceCloudEnabled: prometheusDevice("cloud_service_enabled", "determines if Cloud Service is enabled or not for the current device"),
		DeviceBattery:      prometheusDevice("battery", "The current battery level"),
		LockLockState: prometheus.NewDesc(
			prometheus.BuildFQName(
				namespace,
				"device",
				"lock_state",
			),
			"The lock state of the device",
			[]string{"id", "name", "state"},
			nil,
		),
		LockDoorState: prometheus.NewDesc(
			prometheus.BuildFQName(
				namespace,
				"device",
				"door_state",
			),
			"The door state of the device",
			[]string{"id", "name", "state"},
			nil,
		),
	}
}

// prometheusDevice Device-Specific Describer
func prometheusDevice(metric string, help string) *prometheus.Desc {
	return prometheus.NewDesc(
		prometheus.BuildFQName(
			namespace,
			"device",
			fmt.Sprintf("%s", metric),
		),
		help,
		[]string{"id", "name"},
		nil,
	)
}

// Bool2f64 Converts Boolean to a float64 (True = 1, False = 0)
func Bool2f64(b bool) float64 {
	var i float64
	if b {
		i = 1
	} else {
		i = 0
	}
	return i
}

// StateOK If the device state is in an acceptable state return 1, otherwise return 0
func StateOK(state string) float64 {
	switch state {
	case "locked":
		return 1
	case "unlocked":
		return 0
	case "jammed":
		return 0
	case "closed":
		return 1
	default:
		return 0
	}
}
