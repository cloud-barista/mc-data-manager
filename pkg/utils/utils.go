package utils

import (
	"fmt"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

type Object struct {
	ChecksumAlgorithm []string
	ETag              string
	Key               string
	LastModified      time.Time
	Size              int64
	StorageClass      string
}

type Provider string

const (
	AWS Provider = "aws"
	GCP Provider = "gcp"
	NCP Provider = "ncp"
	OPM Provider = "on-premise"
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

func FileExists(filePath string) bool {
	if fi, err := os.Stat(filePath); os.IsExist(err) {
		return !fi.IsDir()
	}
	return false
}

func LogWirte(logger *logrus.Logger, logLevel, FnName, msg string, err error) {
	if logger != nil {
		switch logLevel {
		case "Info":
			logger.Info(fmt.Sprintf("[%s] %s", FnName, msg))
		case "Error":
			logger.Error(fmt.Sprintf("[%s] %s: %v", FnName, msg, err))
		case "Warn":
			logger.Warn(fmt.Sprintf("[%s] %s: %v", FnName, msg, err))
		}
	}
}
