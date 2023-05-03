package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiskImageWithBackingFile(t *testing.T) {
	backingFileContent := []byte("backing file test")
	diskImageContent := []byte("new disk image content")

	backingFile := filepath.Join(t.TempDir(), "backingfile")
	err := os.WriteFile(backingFile, backingFileContent, 0600)
	require.NoError(t, err)

	diskImage := filepath.Join(t.TempDir(), "diskimage")
	err = NewDiskImageWithBackingFile(diskImage, backingFile)
	require.NoError(t, err)

	data, err := os.ReadFile(diskImage)
	require.NoError(t, err)
	require.Equal(t, data, backingFileContent)

	f, err := os.OpenFile(diskImage, os.O_RDWR|os.O_TRUNC, 0600)
	require.NoError(t, err)
	defer f.Close()
	_, err = f.Write(diskImageContent)
	require.NoError(t, err)

	data, err = os.ReadFile(backingFile)
	require.NoError(t, err)
	require.Equal(t, data, backingFileContent)

	data, err = os.ReadFile(diskImage)
	require.NoError(t, err)
	require.Equal(t, data, diskImageContent)
}
