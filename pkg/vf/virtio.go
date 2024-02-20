package vf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/crc-org/vfkit/pkg/config"
	"github.com/onsi/gocleanup"
	"golang.org/x/sys/unix"

	"github.com/Code-Hex/vz/v3"
	"github.com/pkg/term/termios"
	log "github.com/sirupsen/logrus"
)

type (
	RosettaShare struct {
		*config.RosettaShare
	}
	NVMExpressController struct {
		*config.NVMExpressController
	}
	VirtioBlk struct {
		*config.VirtioBlk
	}
	VirtioFs struct {
		*config.VirtioFs
	}
	VirtioRng struct {
		*config.VirtioRng
	}
	VirtioSerial struct {
		*config.VirtioSerial
		ptyName string
	}
	VirtioVsock struct {
		*config.VirtioVsock
	}
	VirtioInput struct {
		*config.VirtioInput
	}
	VirtioGPU struct {
		*config.VirtioGPU
	}
)

func (dev *NVMExpressController) toVz() (vz.StorageDeviceConfiguration, error) {
	var storageConfig StorageConfig = StorageConfig(dev.StorageConfig)
	attachment, err := storageConfig.toVz()
	if err != nil {
		return nil, err
	}
	devConfig, err := vz.NewNVMExpressControllerDeviceConfiguration(attachment)
	if err != nil {
		return nil, err
	}

	return devConfig, nil
}

func (dev *NVMExpressController) AddToVirtualMachineConfig(vmConfig *VirtualMachineConfiguration) error {
	storageDeviceConfig, err := dev.toVz()
	if err != nil {
		return err
	}
	log.Infof("Adding nvme device (imagePath: %s)", dev.ImagePath)
	vmConfig.storageDevicesConfiguration = append(vmConfig.storageDevicesConfiguration, storageDeviceConfig)

	return nil
}

func (dev *VirtioBlk) toVz() (vz.StorageDeviceConfiguration, error) {
	var storageConfig StorageConfig = StorageConfig(dev.StorageConfig)
	attachment, err := storageConfig.toVz()
	if err != nil {
		return nil, err
	}
	devConfig, err := vz.NewVirtioBlockDeviceConfiguration(attachment)
	if err != nil {
		return nil, err
	}

	if dev.DeviceIdentifier != "" {
		err := devConfig.SetBlockDeviceIdentifier(dev.DeviceIdentifier)
		if err != nil {
			return nil, err
		}
	}

	return devConfig, nil
}

func (dev *VirtioBlk) AddToVirtualMachineConfig(vmConfig *VirtualMachineConfiguration) error {
	storageDeviceConfig, err := dev.toVz()
	if err != nil {
		return err
	}
	log.Infof("Adding virtio-blk device (imagePath: %s)", dev.ImagePath)
	vmConfig.storageDevicesConfiguration = append(vmConfig.storageDevicesConfiguration, storageDeviceConfig)

	return nil
}

func (dev *VirtioInput) toVz() (interface{}, error) {
	var inputConfig interface{}
	if dev.InputType == config.VirtioInputPointingDevice {
		inputConfig, err := vz.NewUSBScreenCoordinatePointingDeviceConfiguration()
		if err != nil {
			return nil, fmt.Errorf("failed to create pointing device configuration: %w", err)
		}

		return inputConfig, nil
	}

	inputConfig, err := vz.NewUSBKeyboardConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to create keyboard device configuration: %w", err)
	}

	return inputConfig, nil
}

func (dev *VirtioInput) AddToVirtualMachineConfig(vmConfig *VirtualMachineConfiguration) error {
	inputDeviceConfig, err := dev.toVz()
	if err != nil {
		return err
	}

	switch conf := inputDeviceConfig.(type) {
	case vz.PointingDeviceConfiguration:
		log.Info("Adding virtio-input pointing device")
		vmConfig.pointingDevicesConfiguration = append(vmConfig.pointingDevicesConfiguration, conf)
	case vz.KeyboardConfiguration:
		log.Info("Adding virtio-input keyboard device")
		vmConfig.keyboardConfiguration = append(vmConfig.keyboardConfiguration, conf)
	}

	return nil
}

func (dev *VirtioGPU) toVz() (vz.GraphicsDeviceConfiguration, error) {
	gpuDeviceConfig, err := vz.NewVirtioGraphicsDeviceConfiguration()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize virtio graphic device: %w", err)
	}
	graphicsScanoutConfig, err := vz.NewVirtioGraphicsScanoutConfiguration(int64(dev.Width), int64(dev.Height))
	if err != nil {
		return nil, fmt.Errorf("failed to create graphics scanout: %w", err)
	}
	gpuDeviceConfig.SetScanouts(
		graphicsScanoutConfig,
	)

	return gpuDeviceConfig, nil
}

func (dev *VirtioGPU) AddToVirtualMachineConfig(vmConfig *VirtualMachineConfiguration) error {
	gpuDeviceConfig, err := dev.toVz()
	if err != nil {
		return err
	}

	log.Infof("Adding virtio-gpu device")

	vmConfig.graphicsDevicesConfiguration = append(vmConfig.graphicsDevicesConfiguration, gpuDeviceConfig)

	return nil
}

func (dev *VirtioFs) toVz() (vz.DirectorySharingDeviceConfiguration, error) {
	if dev.SharedDir == "" {
		return nil, fmt.Errorf("missing mandatory 'sharedDir' option for virtio-fs device")
	}
	var mountTag string
	if dev.MountTag != "" {
		mountTag = dev.MountTag
	} else {
		mountTag = filepath.Base(dev.SharedDir)
	}

	sharedDir, err := vz.NewSharedDirectory(dev.SharedDir, false)
	if err != nil {
		return nil, err
	}
	sharedDirConfig, err := vz.NewSingleDirectoryShare(sharedDir)
	if err != nil {
		return nil, err
	}
	fileSystemDeviceConfig, err := vz.NewVirtioFileSystemDeviceConfiguration(mountTag)
	if err != nil {
		return nil, err
	}
	fileSystemDeviceConfig.SetDirectoryShare(sharedDirConfig)

	return fileSystemDeviceConfig, nil
}

func (dev *VirtioFs) AddToVirtualMachineConfig(vmConfig *VirtualMachineConfiguration) error {
	fileSystemDeviceConfig, err := dev.toVz()
	if err != nil {
		return err
	}
	log.Infof("Adding virtio-fs device")
	vmConfig.directorySharingDevicesConfiguration = append(vmConfig.directorySharingDevicesConfiguration, fileSystemDeviceConfig)
	return nil
}

func (dev *VirtioRng) toVz() (*vz.VirtioEntropyDeviceConfiguration, error) {
	return vz.NewVirtioEntropyDeviceConfiguration()
}

func (dev *VirtioRng) AddToVirtualMachineConfig(vmConfig *VirtualMachineConfiguration) error {
	log.Infof("Adding virtio-rng device")
	entropyConfig, err := dev.toVz()
	if err != nil {
		return err
	}
	vmConfig.entropyDevicesConfiguration = append(vmConfig.entropyDevicesConfiguration, entropyConfig)

	return nil
}

// https://developer.apple.com/documentation/virtualization/running_linux_in_a_virtual_machine?language=objc#:~:text=Configure%20the%20Serial%20Port%20Device%20for%20Standard%20In%20and%20Out
func setRawMode(f *os.File) error {
	var attr unix.Termios
	err := termios.Tcgetattr(f.Fd(), &attr)
	if err != nil {
		return err
	}

	// Put stdin into raw mode, disabling local echo, input canonicalization,
	// and CR-NL mapping.
	attr.Iflag &^= unix.ICRNL
	attr.Lflag &^= unix.ICANON | unix.ECHO

	return termios.Tcsetattr(f.Fd(), termios.TCSANOW, &attr)
}

func (dev *VirtioSerial) toVz() (*vz.VirtioConsoleDeviceSerialPortConfiguration, error) {
	var serialPortAttachment vz.SerialPortAttachment
	var retErr error
	switch {
	case dev.UsesStdio:
		if err := setRawMode(os.Stdin); err != nil {
			return nil, err
		}
		serialPortAttachment, retErr = vz.NewFileHandleSerialPortAttachment(os.Stdin, os.Stdout)
	case dev.UsesPty:
		master, slave, err := termios.Pty()
		if err != nil {
			return nil, err
		}
		// as far as I can tell, we have no use for the slave fd in the
		// vfkit process, the user will open minicom/screen/... /dev/ttys00?
		// when needed
		defer slave.Close()

		// the master fd must stay open for vfkit's lifetime
		gocleanup.Register(func() { _ = master.Close() })

		dev.PtyName = slave.Name()

		if err := setRawMode(master); err != nil {
			return nil, err
		}
		serialPortAttachment, retErr = vz.NewFileHandleSerialPortAttachment(master, master)

	default:
		serialPortAttachment, retErr = vz.NewFileSerialPortAttachment(dev.LogFile, false)
	}
	if retErr != nil {
		return nil, retErr
	}

	return vz.NewVirtioConsoleDeviceSerialPortConfiguration(serialPortAttachment)
}

func (dev *VirtioSerial) AddToVirtualMachineConfig(vmConfig *VirtualMachineConfiguration) error {
	if dev.LogFile != "" {
		log.Infof("Adding virtio-serial device (logFile: %s)", dev.LogFile)
	}
	if dev.UsesStdio {
		log.Infof("Adding stdio console")
	}
	if dev.PtyName != "" {
		return fmt.Errorf("VirtioSerial.PtyName must be empty (current value: %s)", dev.PtyName)
	}

	consoleConfig, err := dev.toVz()
	if err != nil {
		return err
	}
	if dev.UsesPty {
		log.Infof("Using PTY (pty path: %s)", dev.PtyName)
	}
	vmConfig.serialPortsConfiguration = append(vmConfig.serialPortsConfiguration, consoleConfig)

	return nil
}

func (dev *VirtioVsock) AddToVirtualMachineConfig(vmConfig *VirtualMachineConfiguration) error {
	if len(vmConfig.socketDevicesConfiguration) != 0 {
		log.Debugf("virtio-vsock device already present, not adding a second one")
		return nil
	}
	log.Infof("Adding virtio-vsock device")
	vzdev, err := vz.NewVirtioSocketDeviceConfiguration()
	if err != nil {
		return err
	}
	vmConfig.socketDevicesConfiguration = append(vmConfig.socketDevicesConfiguration, vzdev)

	return nil
}

type vfDevice interface {
	config.VirtioDevice
	AddToVirtualMachineConfig(vmConfig *VirtualMachineConfiguration) error
}

func configDevToVfDev(dev config.VirtioDevice) (vfDevice, error) {
	switch d := dev.(type) {
	case *config.USBMassStorage:
		return (*USBMassStorage)(d), nil
	case *config.VirtioBlk:
		return &VirtioBlk{d}, nil
	case *config.RosettaShare:
		return &RosettaShare{d}, nil
	case *config.NVMExpressController:
		return &NVMExpressController{d}, nil
	case *config.VirtioFs:
		return &VirtioFs{d}, nil
	case *config.VirtioNet:
		return &VirtioNet{VirtioNet: d}, nil
	case *config.VirtioRng:
		return &VirtioRng{d}, nil
	case *config.VirtioSerial:
		return &VirtioSerial{VirtioSerial: d}, nil
	case *config.VirtioVsock:
		return &VirtioVsock{d}, nil
	case *config.VirtioInput:
		return &VirtioInput{d}, nil
	case *config.VirtioGPU:
		return &VirtioGPU{d}, nil
	default:
		return nil, fmt.Errorf("Unexpected virtio device type: %T", d)
	}
}

func (config *StorageConfig) toVz() (vz.StorageDeviceAttachment, error) {
	if config.ImagePath == "" {
		return nil, fmt.Errorf("missing mandatory 'path' option for %s device", config.DevName)
	}
	syncMode := vz.DiskImageSynchronizationModeFsync
	caching := vz.DiskImageCachingModeCached
	return vz.NewDiskImageStorageDeviceAttachmentWithCacheAndSync(config.ImagePath, config.ReadOnly, caching, syncMode)
}

func (dev *USBMassStorage) toVz() (vz.StorageDeviceConfiguration, error) {
	var storageConfig StorageConfig = StorageConfig(dev.StorageConfig)
	attachment, err := storageConfig.toVz()
	if err != nil {
		return nil, err
	}
	return vz.NewUSBMassStorageDeviceConfiguration(attachment)
}

func (dev *USBMassStorage) AddToVirtualMachineConfig(vmConfig *VirtualMachineConfiguration) error {
	storageDeviceConfig, err := dev.toVz()
	if err != nil {
		return err
	}
	log.Infof("Adding USB mass storage device (imagePath: %s)", dev.ImagePath)
	vmConfig.storageDevicesConfiguration = append(vmConfig.storageDevicesConfiguration, storageDeviceConfig)

	return nil
}

type StorageConfig config.StorageConfig

type USBMassStorage config.USBMassStorage
