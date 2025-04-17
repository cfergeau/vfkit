package cloudinit

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/kdomanski/iso9660"
	log "github.com/sirupsen/logrus"
)

// it generates a cloud init image by taking the files passed by the user
// as cloud-init expects files with a specific name (e.g user-data, meta-data) we check the filenames to retrieve the correct info
// if some file is not passed by the user, an empty file will be copied to the cloud-init ISO to
// guarantee it to work (user-data and meta-data files are mandatory and both must exists, even if they are empty)
// if both files are missing it returns an error
func GenerateISOFromFiles(files []string) (string, error) {
	if len(files) == 0 {
		return "", nil
	}

	configFiles := map[string]io.Reader{
		"user-data": nil,
		"meta-data": nil,
	}

	hasConfigFile := false
	for _, path := range files {
		if path == "" {
			continue
		}
		file, err := os.Open(path)
		if err != nil {
			return "", err
		}
		defer file.Close()

		filename := filepath.Base(path)
		if _, ok := configFiles[filename]; ok {
			hasConfigFile = true
			configFiles[filename] = file
		}
	}

	if !hasConfigFile {
		return "", fmt.Errorf("cloud-init needs user-data and meta-data files to work")
	}

	return GenerateISO(configFiles)
}

// It generates a temp ISO file containing the files passed by the user
// It also register an exit handler to delete the file when vfkit exits
func GenerateISO(files map[string]io.Reader) (string, error) {
	writer, err := iso9660.NewWriter()
	if err != nil {
		return "", fmt.Errorf("failed to create writer: %w", err)
	}

	defer func() {
		if err := writer.Cleanup(); err != nil {
			log.Error(err)
		}
	}()

	for name, reader := range files {
		// if reader is nil, we set it to an empty file
		if reader == nil {
			reader = bytes.NewReader([]byte{})
		}
		err = writer.AddFile(reader, name)
		if err != nil {
			return "", fmt.Errorf("failed to add %s file: %w", name, err)
		}
	}

	isoFile, err := os.CreateTemp("", "vfkit-cloudinit-")
	if err != nil {
		return "", fmt.Errorf("unable to create temporary cloud-init ISO file: %w", err)
	}

	defer func() {
		if err := isoFile.Close(); err != nil {
			log.Error(fmt.Errorf("failed to close cloud-init ISO file: %w", err))
		}
	}()

	err = writer.WriteTo(isoFile, "cidata")
	if err != nil {
		os.Remove(isoFile.Name())
		return "", fmt.Errorf("failed to write cloud-init ISO image: %w", err)
	}

	return isoFile.Name(), nil
}
