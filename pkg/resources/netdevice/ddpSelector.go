package resources

import (
	"github.com/intel/sriov-network-device-plugin/pkg/types"
	"github.com/intel/sriov-network-device-plugin/pkg/utils"
)

type ddpSelector struct {
	profiles []string
}

// newDdpSelector returns a DeviceSelector interface to filter devices based on available DDP profile
func newDdpSelector(profiles []string) NetDeviceSelector {
	return &ddpSelector{profiles: profiles}
}

func (ds *ddpSelector) Filter(inDevices []types.GenericPciDevice) []types.GenericPciDevice {
	filteredList := make([]types.GenericPciDevice, 0)

	for _, dev := range inDevices {
		netDev := dev.(PciNetDevice)
		ddpProfile := netDev.GetDDPProfiles()
		if ddpProfile != "" && utils.Contains(ds.profiles, ddpProfile) {
			filteredList = append(filteredList, dev)
		}
	}

	return filteredList
}
