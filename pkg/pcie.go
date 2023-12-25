package node

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/gravitational/trace"
)

const (
	VendorId      = 0x4586
	DeviceId      = 0x1234 // TODO
	sysfsPcieRoot = "/sys/bus/pci/devices/"
)

// TODO eventually replace this with https://github.com/TimRots/gutil-linux/tree/master
// if Mixtile ever registers with PCI-SIG
type BladePcieEntry struct {
	Address string
}

func (bpe *BladePcieEntry) GetPath() string {
	return filepath.Join(sysfsPcieRoot, bpe.Address)
}

func (bpe *BladePcieEntry) String() string {
	// This is hard coded because the vendor (Mixtile) does not appear to actually be a PCI-SIG member, and
	// is using another org's vendor ID
	return strings.TrimLeft(bpe.Address, "0000:") + " Network controller: Mixtile Limited Blade 3 (rev 01)"
}

type IPCIeController interface {
	RescanAll() error
	List(activeOnly bool) ([]*BladePcieEntry, error)
}

type PCIeController struct {
	pcieRoot string // Used for testing purposes
}

func NewPCIeController() *PCIeController {
	return &PCIeController{
		pcieRoot: sysfsPcieRoot,
	}
}

func (pciec *PCIeController) RescanAll() error {
	err := os.WriteFile(filepath.Join(pciec.pcieRoot, "rescan"), []byte("1\n"), 0)
	return trace.Wrap(err, "failed to trigger PCIe rescan via sysfs interface")
}

func (pciec *PCIeController) List(activeOnly bool) ([]*BladePcieEntry, error) {
	entries := make([]*BladePcieEntry, 0)

	vendorFileString := fmt.Sprintf("%x", VendorId)
	deviceFileString := fmt.Sprintf("%x", DeviceId)
	err := filepath.WalkDir(pciec.pcieRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return trace.Wrap(err, "an error occursed while walking over directory")
		}

		// Skip the root directory itself
		if path == pciec.pcieRoot {
			return nil
		}

		// No checks should be needed to determine if the entry is a PCIe address, as only symlinks named for PCIe
		// addresses should exist in this directory.

		// Reject entries for other vendors and devices
		data, err := os.ReadFile(filepath.Join(path, "vendor"))
		if err != nil {
			return trace.Wrap(err, "failed to read vendor ID file for device at %q", path)
		}
		if strings.TrimSpace(string(data)) != vendorFileString {
			return nil
		}

		data, err = os.ReadFile(filepath.Join(path, "device"))
		if err != nil {
			return trace.Wrap(err, "failed to read device ID file for device at %q", path)
		}
		if strings.TrimSpace(string(data)) != deviceFileString {
			return nil
		}

		// Reject disabled/offline devices
		if activeOnly {
			data, err := os.ReadFile(filepath.Join(path, "enable"))
			if err != nil {
				return trace.Wrap(err, "failed to determine if device at %q is enabled", path)
			}

			if strings.TrimSpace(string(data)) != "1" {
				return nil
			}
		}

		entries = append(entries, &BladePcieEntry{
			Address: d.Name(),
		})

		return nil
	})

	return entries, trace.Wrap(err, "failed to walk over PCIe root directory %q", pciec.pcieRoot)
}
