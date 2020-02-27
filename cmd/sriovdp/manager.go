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

package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	//"strconv"

	"github.com/golang/glog"
	"github.com/jaypipes/ghw"

	"github.com/intel/sriov-network-device-plugin/pkg/resources"
	"github.com/intel/sriov-network-device-plugin/pkg/types"
	"github.com/intel/sriov-network-device-plugin/pkg/utils"
)

const (
	socketSuffix = "sock"
)

/*
Network controller subclasses. ref: https://pci-ids.ucw.cz/read/PD/02
		00	Ethernet controller
		01	Token ring network controller
		02	FDDI network controller
		03	ATM network controller
		04	ISDN controller
		05	WorldFip controller
		06	PICMG controller
		07	Infiniband controller
		08	Fabric controller
		80	Network controller
*/

type cliParams struct {
	configFile     string
	resourcePrefix string
}

type resourceManager struct {
	cliParams
	pluginWatchMode bool
	socketSuffix    string
	rFactory        types.ResourceFactory
	configList      []*types.ResourceConfig // resourceName -> resourcePool
	resourceServers []types.ResourceServer
	deviceList      map[types.DeviceType][]types.GenericPciDevice // all devices in host
	linkWatchList   map[string]types.LinkWatcher                  // SRIOV PF list - for watching link status
}

func newResourceManager(cp *cliParams) *resourceManager {
	pluginWatchMode := utils.DetectPluginWatchMode(types.SockDir)
	if pluginWatchMode {
		glog.Infof("Using Kubelet Plugin Registry Mode")
	} else {
		glog.Infof("Using Deprecated Device Plugin Registry Path")
	}
	return &resourceManager{
		cliParams:       *cp,
		pluginWatchMode: pluginWatchMode,
		rFactory:        resources.NewResourceFactory(cp.resourcePrefix, socketSuffix, pluginWatchMode),
		deviceList:      make(map[types.DeviceType][]types.GenericPciDevice, 0),
		linkWatchList:   make(map[string]types.LinkWatcher, 0),
	}
}

// Read and validate configurations from Config file
func (rm *resourceManager) readConfig() error {

	res := &types.ResourceConfList{}
	rawBytes, err := ioutil.ReadFile(rm.configFile)

	if err != nil {
		return fmt.Errorf("error reading file %s, %v", rm.configFile, err)

	}

	if err = json.Unmarshal(rawBytes, res); err != nil {
		return fmt.Errorf("error unmarshalling raw bytes %v", err)
	}

	glog.Infof("ResourceList: %+v", res.ResourceList)
	for i := range res.ResourceList {
		var config types.ResourceConfig
		if err = json.Unmarshal(res.ResourceList[i], &config.CommonConfig); err != nil {
			return fmt.Errorf("error unmarshalling common config %v", err)
		}
		switch config.CommonConfig.ResourceType {
		case "netdevice":
			config.DeviceConfig = new(resources.NetDevResourceConfig)
			err := json.Unmarshal(res.ResourceList[i], &config.DeviceConfig)
			if err != nil {
				return fmt.Errorf("error unmarshalling netdev config %v", err)
			}
		}
		rm.configList = append(rm.configList, &config)
	}

	return nil
}

func (rm *resourceManager) initServers() error {
	rf := rm.rFactory
	glog.Infof("number of config: %d\n", len(rm.configList))
	for _, rc := range rm.configList {
		// Create new ResourcePool
		glog.Infof("")
		glog.Infof("Creating new ResourcePool: %s", rc.CommonConfig.ResourceName)
		// Filter once with generic filters:
		filteredDevices := rm.getFilteredDevices(rc)
		rPool, err := rm.rFactory.GetResourcePool(rc, filteredDevices)
		// Inside getresourcePool the devices should be filtered again
		// this time using per-resource filters

		if err != nil {
			glog.Errorf("initServers(): error creating ResourcePool with config %+v: %q", rc, err)
			return err
		}

		// Create ResourceServer with this ResourcePool
		s, err := rf.GetResourceServer(rPool)
		if err != nil {
			glog.Errorf("initServers(): error creating ResourceServer: %v", err)
			return err
		}
		glog.Infof("New resource server is created for %s ResourcePool", rc.CommonConfig.ResourceName)
		rm.resourceServers = append(rm.resourceServers, s)
	}
	return nil
}

func (rm *resourceManager) startAllServers() error {
	for _, rs := range rm.resourceServers {
		if err := rs.Start(); err != nil {
			return err
		}

		// start watcher
		if !rm.pluginWatchMode {
			go rs.Watch()
		}
	}
	return nil
}

func (rm *resourceManager) stopAllServers() error {
	for _, rs := range rm.resourceServers {
		if err := rs.Stop(); err != nil {
			return err
		}
	}
	return nil
}

// Validate configurations. TODO: Add a devicetype-specific validation function?
func (rm *resourceManager) validConfigs() bool {
	resourceNames := make(map[string]string) // resource names placeholder

	for _, conf := range rm.configList {
		// check if name contains acceptable characters
		if !utils.ValidResourceName(conf.CommonConfig.ResourceName) {
			glog.Errorf("resource name \"%s\" contains invalid characters", conf.CommonConfig.ResourceName)
			return false
		}

		// resourcePrefix might be overriden for a given resource pool
		resourcePrefix := rm.cliParams.resourcePrefix
		if conf.CommonConfig.ResourcePrefix != "" {
			resourcePrefix = conf.CommonConfig.ResourcePrefix
		}

		resourceName := resourcePrefix + "/" + conf.CommonConfig.ResourceName

		glog.Infof("validating resource name \"%s\"", resourceName)

		// ensure that resource name is unique
		if _, exists := resourceNames[resourceName]; exists {
			// resource name already exist
			glog.Errorf("resource name \"%s\" already exists", resourceName)
			return false
		}

		resourceNames[resourceName] = resourceName
	}

	return true
}

var deviceTypes = []types.DeviceType{
	resources.NetDeviceType{},
	//OtherDeviceType{},
}

func (rm *resourceManager) discoverHostDevices() error {
	glog.Infoln("discovering host network devices")

	pci, err := ghw.PCI()
	if err != nil {
		return fmt.Errorf("discoverDevices(): error getting PCI info: %v", err)
	}

	devices := pci.ListDevices()
	if len(devices) == 0 {
		glog.Warningf("discoverDevices(): no PCI devices found")
	}

	for _, dtype := range deviceTypes {
		/// DO PER DEVICE DISCOVERY
		deviceList, watchList := dtype.DiscoverHostDevices(devices, rm.rFactory)
		rm.deviceList[dtype] = deviceList
		for _, addr := range watchList {
			rm.addToLinkWatchList(addr)
		}
	}
	return nil
}

func (rm *resourceManager) addToLinkWatchList(pciAddr string) {
	netNames, err := utils.GetNetNames(pciAddr)
	if err == nil {
		// There are some cases, where we may get multiple netdevice name for a PCI device
		// Only add one device
		if len(netNames) > 0 {
			netName := netNames[0]
			lw := &linkWatcher{ifName: netName}
			if _, ok := rm.linkWatchList[pciAddr]; !ok {
				rm.linkWatchList[netName] = lw
				glog.Infof("%s added to linkWatchList", netName)
			}
		}
	}
}

func findResourcetype(typename string) types.DeviceType {
	for _, dtype := range deviceTypes {
		if dtype.GetName() == typename {
			return dtype
		}
	}
	return nil
}

// applyFilters returned a subset PciNetDevices by applying given selectors values in the following orders:
// "vendors", "devices", "drivers", "pfNames", "ddpProfiles".
// Each selector gets a new sub-set of devices from the result of previous one.
func (rm *resourceManager) getFilteredDevices(rc *types.ResourceConfig) []types.GenericPciDevice {
	dtype := findResourcetype(rc.CommonConfig.ResourceType)
	var selectors []string
	if dtype == nil {
		return nil // TODO ERROR!
	}
	filteredDevice := rm.deviceList[dtype]

	// Run common selectors

	rf := rm.rFactory
	// filter by vendor list
	selectors = rc.DeviceConfig.GetSelector("vendors")
	if selectors != nil && len(selectors) > 0 {
		if selector, err := rf.GetSelector("vendors", selectors); err == nil {
			filteredDevice = selector.Filter(filteredDevice)
		}
	}

	// filter by device list

	selectors = rc.DeviceConfig.GetSelector("devices")
	if selectors != nil && len(selectors) > 0 {
		if selector, err := rf.GetSelector("devices", selectors); err == nil {
			filteredDevice = selector.Filter(filteredDevice)
		}
	}

	// filter by driver list
	selectors = rc.DeviceConfig.GetSelector("drivers")
	if selectors != nil && len(selectors) > 0 {
		if selector, err := rf.GetSelector("drivers", selectors); err == nil {
			filteredDevice = selector.Filter(filteredDevice)
		}
	}

	// Run device specific selectors

	filteredDevice = dtype.FilterDevices(rc, rf, filteredDevice)
	return filteredDevice
}

type linkWatcher struct {
	ifName string
	// subscribers []LinkSubscriber
}

func (lw *linkWatcher) Subscribe() {

}
