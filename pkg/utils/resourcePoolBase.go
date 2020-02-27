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

package utils

import (
	"github.com/golang/glog"
	"github.com/intel/sriov-network-device-plugin/pkg/types"
	pluginapi "k8s.io/kubernetes/pkg/kubelet/apis/deviceplugin/v1beta1"
)

// ResourcePoolBase serves as a base class for other ResourceType-specific
// ResourcePools. It implements the ResourcePool Interface
type ResourcePoolBase struct {
	Config     *types.ResourceConfig
	Devices    map[string]*pluginapi.Device
	DevicePool map[string]types.GenericPciDevice
}

//var _ types.ResourcePool = &ResourcePoolBase{} why???

// GetConfig returns the configuration object
func (rp *ResourcePoolBase) GetConfig() *types.ResourceConfig {
	return rp.Config
}

// InitDevice initializes the device
func (rp *ResourcePoolBase) InitDevice() error {
	// Not implemented
	return nil
}

// GetResourceName returns the name of the resource pool
func (rp *ResourcePoolBase) GetResourceName() string {
	return rp.Config.CommonConfig.ResourceName
}

// GetResourcePrefix returns the prefix of the resource pool
func (rp *ResourcePoolBase) GetResourcePrefix() string {
	return rp.Config.CommonConfig.ResourcePrefix
}

// GetDevices returns the list of devices in this pool
func (rp *ResourcePoolBase) GetDevices() map[string]*pluginapi.Device {
	// returns all devices from devices[]
	return rp.Devices
}

// Probe probes the devices
func (rp *ResourcePoolBase) Probe() bool {
	// TO-DO: Implement this
	return false
}

// GetDeviceSpecs returns the devicespecs that match a deviceID array
func (rp *ResourcePoolBase) GetDeviceSpecs(deviceIDs []string) []*pluginapi.DeviceSpec {
	glog.Infof("GetDeviceSpecs(): for devices: %v", deviceIDs)
	devSpecs := make([]*pluginapi.DeviceSpec, 0)

	for _, id := range deviceIDs {
		if dev, ok := rp.DevicePool[id]; ok {
			newSpecs := dev.GetDeviceSpecs()
			for _, ds := range newSpecs {
				if !DeviceSpecExist(devSpecs, ds) {
					devSpecs = append(devSpecs, ds)
				}

			}

		}
	}
	return devSpecs
}

// GetEnvs returns the Env values that match a deviceID array
func (rp *ResourcePoolBase) GetEnvs(deviceIDs []string) []string {
	glog.Infof("GetEnvs(): for devices: %v", deviceIDs)
	devEnvs := make([]string, 0)

	// Consolidates all Envs
	for _, id := range deviceIDs {
		if dev, ok := rp.DevicePool[id]; ok {
			env := dev.GetEnvVal()
			devEnvs = append(devEnvs, env)
		}
	}

	return devEnvs
}

// GetMounts returns the mount definitions
func (rp *ResourcePoolBase) GetMounts(deviceIDs []string) []*pluginapi.Mount {
	glog.Infof("GetMounts(): for devices: %v", deviceIDs)
	devMounts := make([]*pluginapi.Mount, 0)

	for _, id := range deviceIDs {
		if dev, ok := rp.DevicePool[id]; ok {
			mnt := dev.GetMounts()
			devMounts = append(devMounts, mnt...)
		}
	}
	return devMounts
}

// DeviceSpecExist returns whether a device device specs exists in an array of them
func DeviceSpecExist(specs []*pluginapi.DeviceSpec, newSpec *pluginapi.DeviceSpec) bool {
	for _, sp := range specs {
		if sp.HostPath == newSpec.HostPath {
			return true
		}
	}
	return false
}
