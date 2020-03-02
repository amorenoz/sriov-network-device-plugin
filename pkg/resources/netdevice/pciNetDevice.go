package resources

import (
	"fmt"
	"strconv"

	"github.com/golang/glog"
	"github.com/intel/sriov-network-device-plugin/pkg/types"
	"github.com/intel/sriov-network-device-plugin/pkg/utils"
	"github.com/jaypipes/ghw"
	"github.com/vishvananda/netlink"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

const (
	netClass = 0x02 // Device class - Network controller.	 ref: https://pci-ids.ucw.cz/read/PD/02 (for Sub-Classes)
)

// PciNetDevice extends GenericPciDevice
type PciNetDevice interface {
	types.GenericPciDevice
	GetPFName() string
	GetNetName() string
	GetLinkSpeed() string
	GetLinkType() string
	GetRdmaSpec() types.RdmaSpec
	GetDDPProfiles() string
}

type pciNetDevice struct {
	utils.PciDeviceBase
	rdmaSpec  types.RdmaSpec
	linkType  string
	linkSpeed string
	pfName    string
}

// NetDeviceSelector provides an interface for filtering a list of network devices
type NetDeviceSelector interface {
	types.DeviceSelector
}

// NetDevResourceConfig is the PciNetDevice resource configuration
type NetDevResourceConfig struct {
	IsRdma    bool // the resource support rdma
	Selectors struct {
		types.CommonSelectors
		PfNames     []string `json:"pfNames,omitempty"`
		LinkTypes   []string `json:"linkTypes,omitempty"`
		DDPProfiles []string `json:"ddpProfiles,omitempty"`
	} `json:"selectors,omitempty"` // Whether devices have SRIOV virtual function capabilities or not
}

// GetSelector returns the selector by type
func (rc NetDevResourceConfig) GetSelector(sel string) []string {
	switch sel {
	case "vendors":
		return rc.Selectors.Vendors
	case "devices":
		return rc.Selectors.Devices
	case "drivers":
		return rc.Selectors.Devices
	}
	return nil
}

// NetDeviceType implementes DeviceType interface
type NetDeviceType struct{}

// GetName returns the name of the DeviceType
func (n NetDeviceType) GetName() string {
	return "netdevice"
}

// DiscoverHostDevices discovers Host devices of type NetDeviceType
func (NetDeviceType) DiscoverHostDevices(pciDevs []*ghw.PCIDevice, rFactory types.ResourceFactory) ([]types.GenericPciDevice, []string) {

	watchList := make([]string, 0)
	devList := make([]types.GenericPciDevice, 0)

	for _, device := range pciDevs {
		devClass, err := strconv.ParseInt(device.Class.ID, 16, 64)
		if err != nil {
			glog.Warningf("discoverDevices(): unable to parse device class for device %+v %q", device, err)
			continue
		}
		// only interested in network class
		if devClass == netClass {
			vendor := device.Vendor
			vendorName := vendor.Name
			if len(vendor.Name) > 20 {
				vendorName = string([]byte(vendorName)[0:17]) + "..."
			}
			product := device.Product
			productName := product.Name
			if len(product.Name) > 40 {
				productName = string([]byte(productName)[0:37]) + "..."
			}
			glog.Infof("discoverDevices(): device found: %-12s\t%-12s\t%-20s\t%-40s", device.Address, device.Class.ID, vendorName, productName)

			// exclude device in-use in host
			if isDefaultRoute, _ := hasDefaultRoute(device.Address); !isDefaultRoute {

				aPF := utils.IsSriovPF(device.Address)
				aVF := utils.IsSriovVF(device.Address)

				if aPF || !aVF {
					// add to linkWatchList
					watchList = append(watchList, device.Address)
				}

				if aPF && utils.SriovConfigured(device.Address) {
					// do not add this device in net device list
					continue
				}

				if newDevice, err := NewPciNetDevice(device, rFactory); err == nil {
					devList = append(devList, newDevice)
				} else {
					glog.Errorf("discoverDevices() error adding new device: %q", err)
				}
			}

		}
	}
	return devList, watchList
}

// FilterDevices filters devices of type NetDeviceType depending on pool configuration
func (NetDeviceType) FilterDevices(rc *types.ResourceConfig, rf types.ResourceFactory, fd []types.GenericPciDevice) []types.GenericPciDevice {

	filteredDevice := fd
	pc := rc.DeviceConfig.(NetDevResourceConfig)
	// filter by PfNames list
	if pc.Selectors.PfNames != nil && len(pc.Selectors.PfNames) > 0 {
		if selector := newPfNameSelector(pc.Selectors.PfNames); selector != nil {
			filteredDevice = selector.Filter(filteredDevice)
		}
	}

	// filter by linkTypes list
	if pc.Selectors.LinkTypes != nil && len(pc.Selectors.LinkTypes) > 0 {
		if len(pc.Selectors.LinkTypes) > 1 {
			glog.Warningf("Link type selector should have a single value.")
		}
		if selector := newLinkTypeSelector(pc.Selectors.LinkTypes); selector != nil {
			filteredDevice = selector.Filter(filteredDevice)
		}
	}

	// filter by DDP Profiles list
	if pc.Selectors.DDPProfiles != nil && len(pc.Selectors.DDPProfiles) > 0 {
		if selector := newDdpSelector(pc.Selectors.DDPProfiles); selector != nil {
			filteredDevice = selector.Filter(filteredDevice)
		}
	}

	// filter for rdma devices
	if pc.IsRdma {
		rdmaDevices := make([]types.GenericPciDevice, 0)
		for _, dev := range filteredDevice {
			pciDev := dev.(PciNetDevice)
			if pciDev.GetRdmaSpec().IsRdma() {
				rdmaDevices = append(rdmaDevices, pciDev)
			}
		}
		filteredDevice = rdmaDevices
	}
	return filteredDevice

}

// hasDefaultRoute returns true if PCI network device is default route interface
func hasDefaultRoute(pciAddr string) (bool, error) {

	// inUse := false
	// Get net interface name
	ifNames, err := utils.GetNetNames(pciAddr)
	if err != nil {
		return false, fmt.Errorf("error trying get net device name for device %s", pciAddr)
	}

	if len(ifNames) > 0 { // there's at least one interface name found
		for _, ifName := range ifNames {
			link, err := netlink.LinkByName(ifName)
			if err != nil {
				glog.Errorf("expected to get valid host interface with name %s: %q", ifName, err)
			}

			routes, err := netlink.RouteList(link, netlink.FAMILY_V4) // IPv6 routes: all interface has at least one link local route entry
			for _, r := range routes {
				if r.Dst == nil {
					glog.Infof("excluding interface %s:  default route found: %+v", ifName, r)
					return true, nil
				}
			}
		}
	}

	return false, nil
}

// Convert NUMA node number to string.
// A node of -1 represents "unknown" and is converted to the empty string.
func nodeToStr(nodeNum int) string {
	if nodeNum >= 0 {
		return strconv.Itoa(nodeNum)
	}
	return ""
}

// NewPciNetDevice returns an instance of PciNetDevice interface
func NewPciNetDevice(pciDevice *ghw.PCIDevice, rFactory types.ResourceFactory) (PciNetDevice, error) {
	///	// populate all fields in pciNetDevice here

	// 			1. get PF details, add PF info in its pciNetDevice struct
	// 			2. Get driver info
	var ifName string
	pciAddr := pciDevice.Address
	driverName, err := utils.GetDriverName(pciAddr)
	if err != nil {
		return nil, err
	}

	netDevs, _ := utils.GetNetNames(pciAddr)
	if len(netDevs) == 0 {
		ifName = ""
	} else {
		ifName = netDevs[0]
	}
	pfName, err := utils.GetPfName(pciAddr)

	if err != nil {
		glog.Warningf("unable to get PF name %q", err.Error())
	}
	vfID, err := utils.GetVFID(pciAddr)
	if err != nil {
		return nil, err
	}

	// 			3. Get Device file info (e.g., uio, vfio specific)
	// Get DeviceInfoProvider using device driver
	infoProvider := rFactory.GetInfoProvider(driverName)
	dSpecs := infoProvider.GetDeviceSpecs(pciAddr)
	mnt := infoProvider.GetMounts(pciAddr)
	env := infoProvider.GetEnvVal(pciAddr)
	rdmaSpec := rFactory.GetRdmaSpec(pciDevice.Address)
	nodeNum := utils.GetDevNode(pciAddr)
	apiDevice := &pluginapi.Device{
		ID:     pciAddr,
		Health: pluginapi.Healthy,
	}
	if nodeNum >= 0 {
		numaInfo := &pluginapi.NUMANode{
			ID: int64(nodeNum),
		}
		apiDevice.Topology = &pluginapi.TopologyInfo{
			Nodes: []*pluginapi.NUMANode{numaInfo},
		}
	}

	linkType := ""
	if len(ifName) > 0 {
		la, err := utils.GetLinkAttrs(ifName)
		if err != nil {
			return nil, err
		}
		linkType = la.EncapType
	}

	// 			4. Create pciNetDevice object with all relevent info
	return &pciNetDevice{
		PciDeviceBase: utils.PciDeviceBase{
			PciDevice:   pciDevice,
			IfName:      ifName,
			Driver:      driverName,
			VfID:        vfID,
			APIDevice:   apiDevice,
			DeviceSpecs: dSpecs,
			Mounts:      mnt,
			Env:         env,
			Numa:        nodeToStr(nodeNum),
		},
		rdmaSpec:  rdmaSpec,
		linkType:  linkType,
		linkSpeed: "", // TO-DO: Get this using utils pkg
		pfName:    pfName,
	}, nil
}

func (nd *pciNetDevice) GetPFName() string {
	return nd.pfName
}

func (nd *pciNetDevice) GetNetName() string {
	return nd.IfName
}

func (nd *pciNetDevice) GetLinkSpeed() string {
	return nd.linkSpeed
}

func (nd *pciNetDevice) GetRdmaSpec() types.RdmaSpec {
	return nd.rdmaSpec
}

func getPFInfos(pciAddr string) (pfAddr, pfName string, err error) {
	return "", "", nil
}

func (nd *pciNetDevice) GetLinkType() string {
	return nd.linkType
}

func (nd *pciNetDevice) GetDDPProfiles() string {
	ddpProfile, err := utils.GetDDPProfiles(nd.PciDevice.Address)
	if err != nil {
		glog.Infof("GetDDPProfiles(): unable to get ddp profiles for device %s : %q", nd.PciDevice.Address, err)
		return ""
	}
	return ddpProfile
}

// NetDevResourcePool extends the ResourcePool with device-specific logic
type NetDevResourcePool struct {
	utils.ResourcePoolBase
}

// GetDeviceSpecs returns the device specs of a netdev resource
// It overrides the default implementation to provide also the rdma char devices
func (rp *NetDevResourcePool) GetDeviceSpecs(deviceIDs []string) []*pluginapi.DeviceSpec {
	glog.Infof("GetDeviceSpecs(): for devices: %v", deviceIDs)
	devSpecs := make([]*pluginapi.DeviceSpec, 0)
	netDevConf := rp.Config.DeviceConfig.(NetDevResourceConfig)

	// Add vfio group specific devices
	for _, id := range deviceIDs {
		if dev, ok := rp.DevicePool[id]; ok {
			netDev := dev.(PciNetDevice)
			newSpecs := netDev.GetDeviceSpecs()
			rdmaSpec := netDev.GetRdmaSpec()
			if netDevConf.IsRdma {
				if rdmaSpec.IsRdma() {
					rdmaDeviceSpec := rdmaSpec.GetRdmaDeviceSpec()
					newSpecs = append(newSpecs, rdmaDeviceSpec...)
				} else {
					glog.Errorf("GetDeviceSpecs(): rdma is required in the configuration but the device %v is not rdma device", id)
				}
			}
			for _, ds := range newSpecs {
				if !utils.DeviceSpecExist(devSpecs, ds) {
					devSpecs = append(devSpecs, ds)
				}

			}

		}
	}
	return devSpecs
}

// NewNetDevResourcePool creates a NetDevResourcePool
func NewNetDevResourcePool(rc *types.ResourceConfig, ad map[string]*pluginapi.Device, dp map[string]types.GenericPciDevice) *NetDevResourcePool {

	return &NetDevResourcePool{
		utils.ResourcePoolBase{
			Config:     rc,
			Devices:    ad,
			DevicePool: dp,
		},
	}
}
