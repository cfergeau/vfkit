package util

import (
	"golang.org/x/sys/unix"
)

// NewDiskImageWithBackingFile creates a new disk image at path using the file
// at backingFilePath as its backing file. The content of the disk
// image will initially be the same as the backing file. When new data is written
// to the disk image, the content of the backing file is not modified.
// The disk image is created using CloneFile which makes this very efficient,
// disk space is initially shared between the disk image and its backing file,
// only modified blocks are taking up additional space.
func NewDiskImageWithBackingFile(path string, backingFilePath string) error {
	return unix.Clonefile(backingFilePath, path, 0)
}
