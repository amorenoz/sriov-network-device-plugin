package netdevice

import (
	"fmt"
	"github.com/intel/sriov-network-device-plugin/pkg/types"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

const (
	vdpaBasePath = "/var/run/vdpa"
)

// UserInfoProvider implements DeviceInfoProvider for User vDPA devices
type UserInfoProvider struct {
	vdpaType string
}

// GetDeviceSpecs returns the DeviceSpecs of the User vDPA Device
func (ip *UserInfoProvider) GetDeviceSpecs(pciAddr string) []*pluginapi.DeviceSpec {
	devSpecs := make([]*pluginapi.DeviceSpec, 0)
	return devSpecs
}

// GetEnvVal returns the ENV value of the User vDPA Device
func (ip *UserInfoProvider) GetEnvVal(pciAddr string) string {
	return vdpaSocketPath(pciAddr)
}

// GetMounts returns the Mount list of the User vDPA Device
func (ip *UserInfoProvider) GetMounts(pciAddr string) []*pluginapi.Mount {
	var vdpaMount pluginapi.Mount
	mounts := make([]*pluginapi.Mount, 0)

	if ip.vdpaType == "server" {
		// If the interfaces is in server mode we can mount the socket file
		path := vdpaSocketPath(pciAddr)
		vdpaMount = pluginapi.Mount{
			ContainerPath: path,
			HostPath:      path,
			ReadOnly:      false,
		}
	} else if ip.vdpaType == "client" {
		// In client mode we have to mount the directory
		path := vdpaSocketDir(pciAddr)
		vdpaMount = pluginapi.Mount{
			ContainerPath: path,
			HostPath:      path,
			ReadOnly:      false,
		}
	}
	mounts = append(mounts, &vdpaMount)
	return mounts
}

// vdpaSocketPath returns the vdpa socket path of a device
func vdpaSocketPath(pciAddr string) string {
	return fmt.Sprintf("%s/%s", vdpaSocketDir(pciAddr), "vdpa.sock")
}

// vdpaSocketPath returns the vdpa socket path of a device
func vdpaSocketDir(pciAddr string) string {
	return fmt.Sprintf("%s/%s", vdpaBasePath, pciAddr)
}

func newUserVdpaInfoProvider(vdpaType string) types.DeviceInfoProvider {
	return &UserInfoProvider{
		vdpaType: vdpaType,
	}
}
