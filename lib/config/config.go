package config

import "os"

// populate the package variables from the content of the TOML
func New(f *os.File) error {
	bytes, err := PreProcess(f, []*os.File{})
	if err != nil {
		return err
	}

	// populate the package variable, Default
	Default, err = getDefault(bytes)
	if err != nil {
		return err
	}

	return nil
}
