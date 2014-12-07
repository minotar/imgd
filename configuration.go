package main

import (
	"encoding/json"
	"io"
	"os"
)

const (
	// The file we read from
	CONFIG_FILE = "config.json"
	// The example file kept in version control. We'll copy and load from this
	// by default.
	CONFIG_EXAMPLE = "config.example.json"
)

type Configuration struct {
	Address string
}

// Reads the configuration from the config file, copying a config into
// place from the example if one does not yet exist.
func (c *Configuration) load() error {
	err := c.ensureConfigExists()
	if err != nil {
		return err
	}

	file, err := os.Open(CONFIG_FILE)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewDecoder(file).Decode(c)
}

// Creates the config.json if it does not exist.
func (c *Configuration) ensureConfigExists() error {
	if _, err := os.Stat(CONFIG_FILE); os.IsNotExist(err) {
		return copyFile(CONFIG_EXAMPLE, CONFIG_FILE)
	} else {
		return nil
	}
}

// Copies *only the contents* of one file to a new path.
func copyFile(src string, dest string) error {
	original, err := os.Open(src)
	if err != nil {
		return err
	}
	defer original.Close()

	destination, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destination.Close()

	// do the actual work
	_, err = io.Copy(destination, original)
	return err
}
