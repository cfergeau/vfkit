package util

import (
	"golang.org/x/sys/unix"
	"os"
)

func NewDiskImage(path string, size int64) error {
	return nil
}

func NewDiskImageWithBackingFile(path string, backingFilePath string) error {
	return unix.Clonefile(backingFilePath, path, 0)
}

func ResizeDiskImage(path string, newSize int64) error {
	return os.Truncate(path, newSize)
}
