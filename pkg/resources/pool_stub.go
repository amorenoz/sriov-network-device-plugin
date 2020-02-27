// Copyright 2018 Intel Corp. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package resources

import (
	"github.com/golang/glog"
	"github.com/intel/sriov-network-device-plugin/pkg/types"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

// Remove the base struct. Let each device type implement the API
type resourcePool struct {
	config     *types.ResourceConfig
	devices    map[string]*pluginapi.Device
	devicePool map[string]types.GenericPciDevice
}

//var _ types.ResourcePool = &resourcePool{} why???

// newResourcePool returns an instance of resourcePool
func newResourcePool(rc *types.ResourceConfig, apiDevices map[string]*pluginapi.Device, devicePool map[string]types.GenericPciDevice) types.ResourcePool {
	switch rc.CommonConfig.ResourceType {
	case "netdev":
		return &netDevResourcePool{
			resourcePool{
				config:     rc,
				devices:    apiDevices,
				devicePool: devicePool,
			},
		}
	case "other":
		return nil
	}
	return nil
}

func (rp *resourcePool) GetConfig() *types.ResourceConfig {
	return rp.config
}

func (rp *resourcePool) InitDevice() error {
	// Not implemented
	return nil
}

func (rp *resourcePool) GetResourceName() string {
	return rp.config.CommonConfig.ResourceName
}

func (rp *resourcePool) GetResourcePrefix() string {
	return rp.config.CommonConfig.ResourcePrefix
}

func (rp *resourcePool) GetDevices() map[string]*pluginapi.Device {
	// returns all devices from devices[]
	return rp.devices
}

func (rp *resourcePool) Probe() bool {
	// TO-DO: Implement this
	return false
}

func (rp *resourcePool) GetDeviceSpecs(deviceIDs []string) []*pluginapi.DeviceSpec {
	glog.Infof("GetDeviceSpecs(): for devices: %v", deviceIDs)
	devSpecs := make([]*pluginapi.DeviceSpec, 0)

	for _, id := range deviceIDs {
		if dev, ok := rp.devicePool[id]; ok {
			newSpecs := dev.GetDeviceSpecs()
			for _, ds := range newSpecs {
				if !deviceSpecExist(devSpecs, ds) {
					devSpecs = append(devSpecs, ds)
				}

			}

		}
	}
	return devSpecs
}

func (rp *resourcePool) GetEnvs(deviceIDs []string) []string {
	glog.Infof("GetEnvs(): for devices: %v", deviceIDs)
	devEnvs := make([]string, 0)

	// Consolidates all Envs
	for _, id := range deviceIDs {
		if dev, ok := rp.devicePool[id]; ok {
			env := dev.GetEnvVal()
			devEnvs = append(devEnvs, env)
		}
	}

	return devEnvs
}

func (rp *resourcePool) GetMounts(deviceIDs []string) []*pluginapi.Mount {
	glog.Infof("GetMounts(): for devices: %v", deviceIDs)
	devMounts := make([]*pluginapi.Mount, 0)

	for _, id := range deviceIDs {
		if dev, ok := rp.devicePool[id]; ok {
			mnt := dev.GetMounts()
			devMounts = append(devMounts, mnt...)
		}
	}
	return devMounts
}

func deviceSpecExist(specs []*pluginapi.DeviceSpec, newSpec *pluginapi.DeviceSpec) bool {
	for _, sp := range specs {
		if sp.HostPath == newSpec.HostPath {
			return true
		}
	}
	return false
}
