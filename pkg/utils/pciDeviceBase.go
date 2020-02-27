package utils

// pciDeviceBase contains the common pciDevice functionality
// as well as helpers and base classes

import (
	"github.com/jaypipes/ghw"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

// PciDeviceBase has some common PCI data and implements GenericPciDevice
type PciDeviceBase struct {
	PciDevice   *ghw.PCIDevice
	IfName      string
	PfAddr      string
	Driver      string
	Vendor      string
	Product     string
	VfID        int
	Env         string
	Numa        string
	APIDevice   *pluginapi.Device
	DeviceSpecs []*pluginapi.DeviceSpec
	Mounts      []*pluginapi.Mount
}

// GetPfPciAddr returns the PF's PCI address
func (nd *PciDeviceBase) GetPfPciAddr() string {
	return nd.PfAddr
}

// GetVendor retuns the Vendor ID
func (nd *PciDeviceBase) GetVendor() string {
	return nd.PciDevice.Vendor.ID
}

// GetDeviceCode retuns the DeviceCode
func (nd *PciDeviceBase) GetDeviceCode() string {
	return nd.PciDevice.Product.ID
}

// GetPciAddr retuns the device's PCI address
func (nd *PciDeviceBase) GetPciAddr() string {
	return nd.PciDevice.Address
}

// GetDriver returns the driver
func (nd *PciDeviceBase) GetDriver() string {
	return nd.Driver
}

// IsSriovPF returns whether it's a PF
func (nd *PciDeviceBase) IsSriovPF() bool {
	return false
}

// GetSubClass returns the Subclass ID
func (nd *PciDeviceBase) GetSubClass() string {
	return nd.PciDevice.Subclass.ID
}

// GetDeviceSpecs returns the API's DeviceSpec data
func (nd *PciDeviceBase) GetDeviceSpecs() []*pluginapi.DeviceSpec {
	return nd.DeviceSpecs
}

// GetEnvVal Returns the environment variable value
func (nd *PciDeviceBase) GetEnvVal() string {
	return nd.Env
}

// GetMounts retunrs the API's Mount array
func (nd *PciDeviceBase) GetMounts() []*pluginapi.Mount {
	return nd.Mounts
}

// GetAPIDevice returns the API's Device info
func (nd *PciDeviceBase) GetAPIDevice() *pluginapi.Device {
	return nd.APIDevice
}

// GetVFID returns the VF ID
func (nd *PciDeviceBase) GetVFID() int {
	return nd.VfID
}
