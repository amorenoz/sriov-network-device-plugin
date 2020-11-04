/*
Copyright 2020 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package netdevice

import (
	"github.com/golang/glog"

	pluginapi "k8s.io/kubelet/pkg/apis/deviceplugin/v1beta1"

	"github.com/intel/sriov-network-device-plugin/pkg/types"
)

/*
   rdmaInfoProvider provides the RDMA information
*/
type rdmaInfoProvider struct {
	rdmaSpec types.RdmaSpec
}

// NewRdmaInfoProvider returns a new Rdma Information Provider
func NewRdmaInfoProvider(rdmaSpec types.RdmaSpec) types.DeviceInfoProvider {
	return &rdmaInfoProvider{
		rdmaSpec: rdmaSpec,
	}
}

// *****************************************************************
/* DeviceInfoProvider Interface */

func (rip *rdmaInfoProvider) GetDeviceSpecs(pciAddr string) []*pluginapi.DeviceSpec {
	if !rip.rdmaSpec.IsRdma() {
		glog.Warningf("GetDeviceSpecs(): RDMA resources for %s not found. Are RDMA modules loaded?", pciAddr)
		return nil
	}
	return rip.rdmaSpec.GetRdmaDeviceSpec()
}

func (rip *rdmaInfoProvider) GetEnvVal(pciAddr string) string {
	return pciAddr
}

func (rip *rdmaInfoProvider) GetMounts(pciAddr string) []*pluginapi.Mount {
	return nil
}
