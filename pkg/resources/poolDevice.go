package resources

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/intel/sriov-network-device-plugin/pkg/types"
	"github.com/intel/sriov-network-device-plugin/pkg/utils"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

// pciNetPoolDevice implements PoolDevice interface
type pciNetPoolDevice struct {
	pciDev      types.PciNetDevice
	env         string
	apiDevice   *pluginapi.Device
	deviceSpecs []*pluginapi.DeviceSpec
	mounts      []*pluginapi.Mount
}

// NewPciNetPoolDevice creates a pciNetPoolDevice
func newPciNetPoolDevice(pciDev types.PciNetDevice, rc *types.ResourceConfig, rFactory types.ResourceFactory) (*pciNetPoolDevice, error) {

	glog.Infof("Creating PciNetPoolDevice for PciNetDevice: %+v\n", pciDev.GetPciAddr())
	// Get Info provider from the driverName
	infoProvider := rFactory.GetInfoProvider(pciDev.GetDriver())

	env := infoProvider.GetEnvVal(pciDev.GetPciAddr())
	mnt := infoProvider.GetMounts(pciDev.GetPciAddr())

	// Set DeviceSpecs
	dSpecs := infoProvider.GetDeviceSpecs(pciDev.GetPciAddr())

	// Append Rdma Specs only if Rdma is in the pool's config
	if rc.IsRdma {
		rdmaSpec := pciDev.GetRdmaSpec()
		if rdmaSpec.IsRdma() {
			rdmaDeviceSpec := rdmaSpec.GetRdmaDeviceSpec()
			for _, spec := range rdmaDeviceSpec {
				dSpecs = append(dSpecs, spec)
			}
		} else {
			return nil, fmt.Errorf("NewPciNetPoolDevice(): rdma is required in the configuration but the device %v is not rdma device", pciDev.GetPciAddr())
		}
	}

	// Create apiDevice
	apiDevice := &pluginapi.Device{
		ID:     pciDev.GetPciAddr(),
		Health: pluginapi.Healthy,
	}
	nodeNum := utils.GetDevNode(pciDev.GetPciAddr())
	if nodeNum >= 0 {
		numaInfo := &pluginapi.NUMANode{
			ID: int64(nodeNum),
		}
		apiDevice.Topology = &pluginapi.TopologyInfo{
			Nodes: []*pluginapi.NUMANode{numaInfo},
		}
	}
	return &pciNetPoolDevice{
		pciDev:      pciDev,
		env:         env,
		apiDevice:   apiDevice,
		deviceSpecs: dSpecs,
		mounts:      mnt,
	}, nil
}

func (pd *pciNetPoolDevice) GetDeviceSpecs() []*pluginapi.DeviceSpec {
	return pd.deviceSpecs
}

func (pd *pciNetPoolDevice) GetEnvVal() string {
	return pd.env
}

func (pd *pciNetPoolDevice) GetMounts() []*pluginapi.Mount {
	return pd.mounts
}

func (pd *pciNetPoolDevice) GetAPIDevice() *pluginapi.Device {
	return pd.apiDevice
}
