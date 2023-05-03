package util

import (
	"path/filepath"
	"syscall"
	"testing"

	"github.com/containers/common/pkg/strongunits"
	"github.com/stretchr/testify/require"
)

func TestDiskImage(t *testing.T) {
	diskImage := filepath.Join(t.TempDir(), "diskimage")
	size := strongunits.KiB(64)
	err := NewDiskImage(diskImage, size, 0640)
	require.NoError(t, err)

	var st syscall.Stat_t
	err = syscall.Stat(diskImage, &st)
	require.NoError(t, err)
	require.Equal(t, int64(size.ToBytes()), st.Size)
	// file must be sparse
	require.Equal(t, int64(0), st.Blocks)

	size = strongunits.KiB(128)
	err = ResizeDiskImage(diskImage, size)
	require.NoError(t, err)

	err = syscall.Stat(diskImage, &st)
	require.NoError(t, err)
	require.Equal(t, int64(size.ToBytes()), st.Size)
	// file must be sparse
	require.Equal(t, int64(0), st.Blocks)
}
