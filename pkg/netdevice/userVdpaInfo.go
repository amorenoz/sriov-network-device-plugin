package netdevice

import (
	"fmt"
	"github.com/amorenoz/govdpa/pkg/uvdpa"
	"github.com/intel/sriov-network-device-plugin/pkg/types"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
	"os"
	"strings"
)

const (
	vdpaBasePath = "/var/run/uvdpa"
)

// UserInfoProvider implements DeviceInfoProvider for User vDPA devices
type UserInfoProvider struct {
	vdpaType string
}

func newUserVdpaInfoProvider(vdpaType string) types.DeviceInfoProvider {
	return &UserInfoProvider{
		vdpaType: vdpaType,
	}
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

	path := vdpaSocketDir(pciAddr)
	vdpaMount = pluginapi.Mount{
		ContainerPath: path,
		HostPath:      path,
		ReadOnly:      false,
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
	pciDir := pciAddr
	pciDir = strings.Replace(pciDir, ":", "__", -1)
	pciDir = strings.Replace(pciDir, ".", "_", -1)
	return fmt.Sprintf("%s/%s", vdpaBasePath, pciDir)
}

func userVdpaAllocate(device string, vdpaType string) error {
	sockDir := vdpaSocketDir(device)
	if _, err := os.Stat(sockDir); !os.IsNotExist(err) {
		if err := os.RemoveAll(sockDir); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(sockDir, 0600); err != nil {
		return err
	}

	u := uvdpa.NewVdpaClient(false)
	if err := u.Init(); err != nil {
		return err
	}

	return u.Create(uvdpa.VhostIface{
		Device: device,
		Socket: vdpaSocketPath(device),
		Mode:   vdpaType,
	})
}
