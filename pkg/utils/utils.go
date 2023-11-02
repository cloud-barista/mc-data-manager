package utils

import (
	"os"
)

// Distinguish between directory and file or directory
func IsDir(path string) error {
	fInfo, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			err := os.MkdirAll(path, 0755)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	} else {
		if !fInfo.IsDir() {
			return err
		}
	}
	return nil
}
