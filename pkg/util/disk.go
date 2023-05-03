package util

import (
	"os"

	"github.com/containers/common/pkg/strongunits"
)

// NewDiskImage creates a new disk image at path. If the file exists, it is truncated.
// If the file does not exist, it's created with perm. In both cases, the file
// is then resized to size. The resulting file is sparse, which means it does
// not use disk space until it's written to.
func NewDiskImage(path string, size strongunits.StorageUnits, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, perm)
	if err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	return ResizeDiskImage(path, size)
}

// ResizeDiskImage changes the size of the disk image at path to newSize.
// Shrinking the disk image is likely to result in guest corruption. If the
// disk image is grown, the additional space is sparsely allocated, this means
// it's not using storage space until i t's written to.
func ResizeDiskImage(path string, newSize strongunits.StorageUnits) error {
	return os.Truncate(path, int64(newSize.ToBytes()))
}
