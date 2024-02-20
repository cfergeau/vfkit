package vf

import (
	"encoding/json"

	"github.com/crc-org/vfkit/pkg/config"
)

// pkg/vf augments the config.VirtioSerial definition with a `PtyName` field
// When using the `inspect` REST API, we want to have this field in the JSON
// output.
// I did not find a way to reuse the config.VirtioSerial.MarshalJSON code to
// add the `Kind` field, so the field is readded here.
func (dev *VirtioSerial) MarshalJSON() ([]byte, error) {
	type devWithKind struct {
		Kind string `json:"kind"`
		VirtioSerial
	}
	return json.Marshal(devWithKind{
		Kind:         "virtioserial",
		VirtioSerial: *dev,
	})
}
func (dev *VirtioSerial) UnmarshalJSON(rawMsg json.RawMessage) (config.VMComponent, error) {
	return nil, nil
}

// similarly to MarshalJSON, when unmarshalling json for a runtime VM
// configuration, we want the PtyName field to be unmarshalled.
// To achieve this we extend the default unmarshalling implemented in pkg/config
// I'm not  fully sure having this PtyName field only in pkg/vf is the right approach
// given the complications it brings.
// One advantage of not having it in pkg/config is that users can't mistakenly assume
// PtyName can be specified when creating the VM. It's auto-allocated by libc/the kernel,
// and we can't control which name the PTY will get.
func unmarshalVirtioSerial(rawMsg json.RawMessage) (config.VMComponent, error) {
	var dev *VirtioSerial
	if err := json.Unmarshal(rawMsg, &dev); err != nil {
		return nil, err
	}
	if dev.PtyName == "" {
		return dev.VirtioSerial, nil
	}

	return dev, nil
}

func (cfg *VirtualMachineConfiguration) UnmarshalJSON(b []byte) error {
	vmConfig := config.VirtualMachine{}
	serialUnmarshaller := map[string]config.DeviceUnmarshaller{
		"virtioserial": unmarshalVirtioSerial,
	}
	if err := vmConfig.UnmarshalJSONCustom(b, serialUnmarshaller); err != nil {
		return err
	}
	cfg.config = &vmConfig
	return nil
}
