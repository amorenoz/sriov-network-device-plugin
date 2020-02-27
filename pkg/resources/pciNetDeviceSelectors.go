package resources

import (
	"fmt"
	"github.com/intel/sriov-network-device-plugin/pkg/types"
	"strconv"
	"strings"
)

// newPfNameSelector returns a NetDevSelector interface for netDev list
func newPfNameSelector(pfNames []string) NetDeviceSelector {
	return &pfNameSelector{pfNames: pfNames}
}

type pfNameSelector struct {
	pfNames []string
}

// PCI NET SPECIFIC SELECTORS
func (s *pfNameSelector) Filter(inDevices []types.GenericPciDevice) []types.GenericPciDevice {
	filteredList := make([]types.GenericPciDevice, 0)
	for _, dev := range inDevices {
		netDev := dev.(PciNetDevice)
		selector := getItem(s.pfNames, netDev.GetPFName())
		if selector != "" {
			if strings.Contains(selector, "#") {
				// Selector does contain VF index in next format:
				// <PFName>#<VFIndexStart>-<VFIndexEnd>
				// In this case both <VFIndexStart> and <VFIndexEnd>
				// are included in range, for example: "netpf0#3-5"
				// The VFs 3,4 and 5 of the PF 'netpf0' will be included
				// in selector pool
				fields := strings.Split(selector, "#")
				if len(fields) != 2 {
					fmt.Printf("Failed to parse %s PF name selector, probably incorrect separator character usage\n", netDev.GetPFName())
					continue
				}
				entries := strings.Split(fields[1], ",")
				for i := 0; i < len(entries); i++ {
					if strings.Contains(entries[i], "-") {
						rng := strings.Split(entries[i], "-")
						if len(rng) != 2 {
							fmt.Printf("Failed to parse %s PF name selector, probably incorrect range character usage\n", netDev.GetPFName())
							continue
						}
						rngSt, err := strconv.Atoi(rng[0])
						if err != nil {
							fmt.Printf("Failed to parse %s PF name selector, start range is incorrect\n", netDev.GetPFName())
							continue
						}
						rngEnd, err := strconv.Atoi(rng[1])
						if err != nil {
							fmt.Printf("Failed to parse %s PF name selector, end range is incorrect\n", netDev.GetPFName())
							continue
						}
						vfID := netDev.GetVFID()
						if vfID >= rngSt && vfID <= rngEnd {
							filteredList = append(filteredList, netDev)
						}
					} else {
						vfid, err := strconv.Atoi(entries[i])
						if err != nil {
							fmt.Printf("Failed to parse %s PF name selector, index is incorrect\n", netDev.GetPFName())
							continue
						}
						vfID := netDev.GetVFID()
						if vfID == vfid {
							filteredList = append(filteredList, netDev)
						}

					}
				}
			} else {
				filteredList = append(filteredList, dev)
			}
		}
	}

	return filteredList
}

// newLinkTypeSelector returns a interface for netDev list
func newLinkTypeSelector(linkTypes []string) NetDeviceSelector {
	return &linkTypeSelector{linkTypes: linkTypes}
}

type linkTypeSelector struct {
	linkTypes []string
}

func (s *linkTypeSelector) Filter(inDevices []types.GenericPciDevice) []types.GenericPciDevice {
	filteredList := make([]types.GenericPciDevice, 0)
	for _, dev := range inDevices {
		netDev := dev.(PciNetDevice)
		if contains(s.linkTypes, netDev.GetLinkType()) {
			filteredList = append(filteredList, dev)
		}
	}
	return filteredList
}
