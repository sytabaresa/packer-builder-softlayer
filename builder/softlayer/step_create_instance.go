package softlayer

import (
	"fmt"
	"github.com/mitchellh/multistep"
	"github.com/mitchellh/packer/packer"
	"log"
)

type stepCreateInstance struct {
	instanceId string
}

func (self *stepCreateInstance) Run(state multistep.StateBag) multistep.StepAction {
	client := state.Get("client").(*SoftlayerClient)
	config := state.Get("config").(config)
	ui := state.Get("ui").(packer.Ui)

	instanceDefinition := &InstanceType{
		HostName:             config.InstanceName,
		Domain:               config.InstanceDomain,
		Datacenter:           config.DatacenterName,
		Cpus:                 config.InstanceCpu,
		Memory:               config.InstanceMemory,
		HourlyBillingFlag:    true,
		LocalDiskFlag:        true,
		DiskCapacity:         config.InstanceDiskCapacity,
		NetworkSpeed:         config.InstanceNetworkSpeed,
		ProvisioningSshKeyId: state.Get("ssh_key_id").(float64),
		BaseImageId:          config.BaseImageId,
		BaseOsCode:           config.BaseOsCode,
	}
	instanceData, _ := client.CreateInstance(*instanceDefinition)
	state.Put("instance_data", instanceData)
	self.instanceId = instanceData["globalIdentifier"].(string)

	ui.Say(fmt.Sprintf("Created instance '%s'", instanceData["globalIdentifier"].(string)))

	return multistep.ActionContinue
}

func (self *stepCreateInstance) Cleanup(state multistep.StateBag) {
	client := state.Get("client").(*SoftlayerClient)
	config := state.Get("config").(config)
	ui := state.Get("ui").(packer.Ui)

	if self.instanceId == "" {
		return
	}

	ui.Say("Waiting for the instance to have no active transactions before destroying it...")

	// We should wait until the instance is up/have no transactions,
	// since if the instance will have some assigned transactions the destroy API call will fail
	err := client.waitForInstanceReady(self.instanceId, config.StateTimeout)
	if err != nil {
		log.Printf("Error destroying instance: %v", err.Error())
		ui.Error(fmt.Sprintf("Error waiting for instance to become ACTIVE for instance (%s)", self.instanceId))
	}

	ui.Say("Destroying instance...")
	err = client.DestroyInstance(self.instanceId)
	if err != nil {
		log.Printf("Error destroying instance: %v", err.Error())
		ui.Error(fmt.Sprintf("Error cleaning up the instance. Please delete the instance (%s) manually", self.instanceId))
	}
}
