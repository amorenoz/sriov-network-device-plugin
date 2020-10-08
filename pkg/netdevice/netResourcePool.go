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

package netdevice

import (
	"github.com/golang/glog"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"github.com/intel/sriov-network-device-plugin/pkg/resources"
	"github.com/intel/sriov-network-device-plugin/pkg/types"
)

type netResourcePool struct {
	*resources.ResourcePoolImpl
	selectors *types.NetDeviceSelectors
}

var _ types.ResourcePool = &netResourcePool{}

// NewNetResourcePool returns an instance of resourcePool
func NewNetResourcePool(rc *types.ResourceConfig, apiDevices map[string]*pluginapi.Device, devicePool map[string]types.PciDevice) types.ResourcePool {
	rp := resources.NewResourcePool(rc, apiDevices, devicePool)
	s, _ := rc.SelectorObj.(*types.NetDeviceSelectors)
	return &netResourcePool{
		ResourcePoolImpl: rp,
		selectors:        s,
	}
}

// Overrides GetDeviceSpecs
func (rp *netResourcePool) GetDeviceSpecs(deviceIDs []string) []*pluginapi.DeviceSpec {
	glog.Infof("GetDeviceSpecs(): for devices: %v", deviceIDs)
	devSpecs := make([]*pluginapi.DeviceSpec, 0)

	devicePool := rp.GetDevicePool()

	// Add device driver specific and rdma specific devices
	for _, id := range deviceIDs {
		if dev, ok := devicePool[id]; ok {
			netDev := dev.(types.PciNetDevice) // convert generic PciDevice to PciNetDevice
			newSpecs := netDev.GetDeviceSpecs()
			if rp.selectors.NeedVhostNet {
				if VhostNetDeviceExist() {
					vhostNetDeviceSpec := GetVhostNetDeviceSpec()
					newSpecs = append(newSpecs, vhostNetDeviceSpec...)
				} else {
					glog.Errorf("GetDeviceSpecs(): vhost-net is required in the configuration but /dev/vhost-net doesn't exist")
				}
			}
			for _, ds := range newSpecs {
				if !rp.DeviceSpecExist(devSpecs, ds) {
					devSpecs = append(devSpecs, ds)
				}

			}

		}
	}
	return devSpecs
}
