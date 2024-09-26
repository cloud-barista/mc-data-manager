/*
Copyright 2023 The Cloud-Barista Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package utils

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/cloud-barista/mc-data-manager/models"
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

// Enum Validation
func IsValidStatus(s models.Status) bool {
	switch s {
	case models.StatusActive, models.StatusInactive, models.StatusPending, models.StatusFailed, models.StatusCompleted:
		return true
	}
	return false
}

func IsValidServiceType(s models.CloudServiceType) bool {
	switch s {
	case models.ComputeService, models.ObejectStorage, models.RDBMS, models.NRDBMS:
		return true
	}
	return false
}

func IsValidTaskType(s models.TaskType) bool {
	switch s {
	case models.Generate, models.Migrate, models.Backup, models.Restore:
		return true
	}
	return false
}

// GEN ID
func GenerateTaskID(opId string, index int) string {
	return fmt.Sprintf("%s-task-%d-%s", opId, index, time.Now().Format("20060102-150405"))
}

func GenerateFlowID(opId string) string {
	return fmt.Sprintf("%s-flow-%s", opId, time.Now().Format("20060102-150405"))
}

func GenerateScheduleID(opId string) string {
	return fmt.Sprintf("%s-schedule-%s", opId, time.Now().Format("20060102-150405"))
}

// validateCronExpression checks if the provided cron expression is valid.
// It returns an error if the expression is invalid.
func validateCronExpression(cronExpr string) error {
	// Split the cron expression by spaces
	fields := strings.Fields(cronExpr)
	if len(fields) != 5 {
		return errors.New("cron expression must have exactly 5 fields")
	}

	// Define regex patterns for each field
	// This is a simplified version and may need to be expanded for full validation
	fieldPatterns := []string{
		`^(\*|([0-5]?\d)(-[0-5]?\d)?(\/\d+)?)$`, // Minute
		`^(\*|([01]?\d|2[0-3])(\/\d+)?)$`,       // Hour
		`^(\*|([01]?\d|2[0-9]|3[01])(\/\d+)?)$`, // Day of Month
		`^(\*|(1[0-2]|0?[1-9])(\/\d+)?)$`,       // Month
		`^(\*|(0|1|2|3|4|5|6)(\/\d+)?)$`,        // Day of Week
	}

	for i, field := range fields {
		matched, err := regexp.MatchString(fieldPatterns[i], field)
		if err != nil {
			return err
		}
		if !matched {
			return errors.New("invalid cron expression in field " + string(i+1))
		}
	}

	return nil
}
