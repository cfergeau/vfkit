package vf

import (
	"fmt"

	"github.com/Code-Hex/vz/v3"
	"github.com/crc-org/vfkit/pkg/config"
)

type VirtualMachine struct {
	*vz.VirtualMachine
	vfConfig *VirtualMachineConfiguration
}

var PlatformType string

func NewVirtualMachine(vmConfig config.VirtualMachine) (*VirtualMachine, error) {
	vfConfig, err := NewVirtualMachineConfiguration(&vmConfig)
	if err != nil {
		return nil, err
	}

	if _, ok := vmConfig.Bootloader.(*config.MacOSBootloader); ok {
		platformConfig, err := NewMacPlatformConfiguration()

		PlatformType = "macos"

		if err != nil {
			return nil, err
		}

		vfConfig.SetPlatformVirtualMachineConfiguration(platformConfig)
	} else {
		PlatformType = "linux"
	}

	return &VirtualMachine{
		vfConfig: vfConfig,
	}, nil
}

func NewMacPlatformConfiguration() (*vz.MacPlatformConfiguration, error) {
	// var HardwareModelVar =[]byte( "YnBsaXN0MDDTAQIDBAUGXxAZRGF0YVJlcHJlc2VudGF0aW9uVmVyc2lvbl8QD1BsYXRmb3JtVmVyc2lvbl8QEk1pbmltdW1TdXBwb3J0ZWRPUxQAAAAAAAAAAAAAAAAAAAABEAKjBwgIEA0QAAgPKz1SY2VpawAAAAAAAAEBAAAAAAAAAAkAAAAAAAAAAAAAAAAAAABt") // Binary plist with {"DataRepresentationVersion":1,"MinimumSupportedOS":[13,0,0],"PlatformVersion":2}
	var AuxiliaryStorageVar = "/Users/foo/VM.bundle/AuxiliaryStorage"
	var HardwareModelVar = "/Users/foo/VM.bundle/HardwareModel"
	var MachineIdentifierVar = "/Users/foo/VM.bundle/MachineIdentifier"

	hardwareModel, err := vz.NewMacHardwareModelWithDataPath(HardwareModelVar)

	if err != nil {
	    return nil, fmt.Errorf("hardwareModel error: %w", err)
	}

	macAuxiliaryStorage, err := vz.NewMacAuxiliaryStorage(
		AuxiliaryStorageVar,
		vz.WithCreatingMacAuxiliaryStorage(hardwareModel),
	)

	if err != nil {
	    return nil, fmt.Errorf("macAuxiliaryStorage error: %w", err)
	}

	machineIdentifier, err := vz.NewMacMachineIdentifierWithDataPath(
		MachineIdentifierVar,
	)

	if err != nil {
	    return nil, fmt.Errorf("machineIdentifier error: %w", err)
	}

	platformConfig, err := vz.NewMacPlatformConfiguration(
		vz.WithMacAuxiliaryStorage(macAuxiliaryStorage),
		vz.WithMacHardwareModel(hardwareModel),
		vz.WithMacMachineIdentifier(machineIdentifier),
	)

	if err != nil {
		return nil, err
	}

	return platformConfig, nil
}

func (vm *VirtualMachine) Start() error {
	if vm.VirtualMachine == nil {
		if err := vm.toVz(); err != nil {
			return err
		}
	}
	return vm.VirtualMachine.Start()
}

func (vm *VirtualMachine) toVz() error {
	vzVMConfig, err := vm.vfConfig.toVz()
	if err != nil {
		return err
	}
	vzVM, err := vz.NewVirtualMachine(vzVMConfig)
	if err != nil {
		return err
	}
	vm.VirtualMachine = vzVM

	return nil
}

func (vm *VirtualMachine) Config() *config.VirtualMachine {
	return vm.vfConfig.config
}

type VirtualMachineConfiguration struct {
	*vz.VirtualMachineConfiguration                             // wrapper for Objective-C type
	config                               *config.VirtualMachine // go-friendly virtual machine configuration definition
	storageDevicesConfiguration          []vz.StorageDeviceConfiguration
	directorySharingDevicesConfiguration []vz.DirectorySharingDeviceConfiguration
	keyboardConfiguration                []vz.KeyboardConfiguration
	pointingDevicesConfiguration         []vz.PointingDeviceConfiguration
	graphicsDevicesConfiguration         []vz.GraphicsDeviceConfiguration
	networkDevicesConfiguration          []*vz.VirtioNetworkDeviceConfiguration
	entropyDevicesConfiguration          []*vz.VirtioEntropyDeviceConfiguration
	serialPortsConfiguration             []*vz.VirtioConsoleDeviceSerialPortConfiguration
	socketDevicesConfiguration           []vz.SocketDeviceConfiguration
}

func NewVirtualMachineConfiguration(vmConfig *config.VirtualMachine) (*VirtualMachineConfiguration, error) {
	vzBootloader, err := toVzBootloader(vmConfig.Bootloader)
	if err != nil {
		return nil, err
	}

	vzVMConfig, err := vz.NewVirtualMachineConfiguration(vzBootloader, vmConfig.Vcpus, uint64(vmConfig.Memory.ToBytes()))
	if err != nil {
		return nil, err
	}

	return &VirtualMachineConfiguration{
		VirtualMachineConfiguration: vzVMConfig,
		config:                      vmConfig,
	}, nil
}

func (cfg *VirtualMachineConfiguration) toVz() (*vz.VirtualMachineConfiguration, error) {
	for _, dev := range cfg.config.Devices {
		if err := AddToVirtualMachineConfig(cfg, dev); err != nil {
			return nil, err
		}
	}
	cfg.SetStorageDevicesVirtualMachineConfiguration(cfg.storageDevicesConfiguration)
	cfg.SetDirectorySharingDevicesVirtualMachineConfiguration(cfg.directorySharingDevicesConfiguration)
	cfg.SetPointingDevicesVirtualMachineConfiguration(cfg.pointingDevicesConfiguration)
	cfg.SetKeyboardsVirtualMachineConfiguration(cfg.keyboardConfiguration)
	cfg.SetGraphicsDevicesVirtualMachineConfiguration(cfg.graphicsDevicesConfiguration)
	cfg.SetNetworkDevicesVirtualMachineConfiguration(cfg.networkDevicesConfiguration)
	cfg.SetEntropyDevicesVirtualMachineConfiguration(cfg.entropyDevicesConfiguration)
	cfg.SetSerialPortsVirtualMachineConfiguration(cfg.serialPortsConfiguration)
	// len(cfg.socketDevicesConfiguration should be 0 or 1
	// https://developer.apple.com/documentation/virtualization/vzvirtiosocketdeviceconfiguration?language=objc
	cfg.SetSocketDevicesVirtualMachineConfiguration(cfg.socketDevicesConfiguration)

	if cfg.config.Timesync != nil && cfg.config.Timesync.VsockPort != 0 {
		// automatically add the vsock device we'll need for communication over VsockPort
		vsockDev := VirtioVsock{
			Port:   cfg.config.Timesync.VsockPort,
			Listen: false,
		}
		if err := vsockDev.AddToVirtualMachineConfig(cfg); err != nil {
			return nil, err
		}
	}

	valid, err := cfg.Validate()
	if err != nil {
		return nil, err
	}
	if !valid {
		return nil, fmt.Errorf("Invalid virtual machine configuration")
	}

	return cfg.VirtualMachineConfiguration, nil
}
