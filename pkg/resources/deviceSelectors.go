package resources

import (
	"github.com/intel/sriov-network-device-plugin/pkg/types"
	"github.com/intel/sriov-network-device-plugin/pkg/utils"
)

// GENERIC FILTERS

// newVendorSelector returns a DeviceSelector interface for vendor list
func newVendorSelector(vendors []string) types.DeviceSelector {
	return &vendorSelector{vendors: vendors}
}

type vendorSelector struct {
	vendors []string
}

func (s *vendorSelector) Filter(inDevices []types.GenericPciDevice) []types.GenericPciDevice {
	filteredList := make([]types.GenericPciDevice, 0)
	for _, dev := range inDevices {
		devVendor := dev.GetVendor()
		if utils.Contains(s.vendors, devVendor) {
			filteredList = append(filteredList, dev)
		}
	}
	return filteredList
}

// newDeviceSelector returns a DeviceSelector interface for device list
func newDeviceSelector(devices []string) types.DeviceSelector {
	return &deviceSelector{devices: devices}
}

type deviceSelector struct {
	devices []string
}

func (s *deviceSelector) Filter(inDevices []types.GenericPciDevice) []types.GenericPciDevice {
	filteredList := make([]types.GenericPciDevice, 0)
	for _, dev := range inDevices {
		devCode := dev.GetDeviceCode()
		if utils.Contains(s.devices, devCode) {
			filteredList = append(filteredList, dev)
		}
	}
	return filteredList
}

// newDriverSelector returns a DeviceSelector interface for driver list
func newDriverSelector(drivers []string) types.DeviceSelector {
	return &driverSelector{drivers: drivers}
}

type driverSelector struct {
	drivers []string
}

func (s *driverSelector) Filter(inDevices []types.GenericPciDevice) []types.GenericPciDevice {
	filteredList := make([]types.GenericPciDevice, 0)
	for _, dev := range inDevices {
		devDriver := dev.GetDriver()
		if utils.Contains(s.drivers, devDriver) {
			filteredList = append(filteredList, dev)
		}
	}
	return filteredList
}
