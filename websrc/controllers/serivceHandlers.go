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
package controllers

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/cloud-barista/mc-data-manager/models"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// DeleteServiceAndTaskAllHandler godoc
//
//	@ID 			DeleteServiceAndTaskAllHandler
//	@Summary		Delete a Task
//	@Description	Delete an All Service and Task.
//	@Tags			[Service]
//	@Produce		json
//	@Success		200		{object}	models.BasicResponse	"Successfully deleted the All Service"
//	@Failure		404		{object}	models.BasicResponse	"Clear All Task , Failed"
//	@Router			/service/clearAll [delete]
func (tc *TaskController) DeleteServiceAndTaskAllHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "Delete-All-Task", "Delete an existing task", start)
	if err := tc.TaskService.ClearServiceAndTaskAll(); err != nil {
		errStr := "Clear All Task , Failed"
		logger.Error().Msg(errStr)
		return ctx.JSON(http.StatusNotFound, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}
	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// applyResourceHandler godoc
//
//	@ID 			applyResourceHandler
//	@Summary		Apply Resources
//	@Description	Execute the apply.sh script to set up resources.
//	@Tags			[Service]
//	@Produce		json
//	@Success		200	{object}	models.BasicResponse	"Successfully applied resources"
//	@Failure		500	{object}	models.BasicResponse	"Failed to apply resources"
//	@Router			/service/apply [post]
func (tc *TaskController) ApplyResourceHandler(ctx echo.Context) error {

	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "apply-resource", "Apply an Data Resource", start)
	// Get the directory of the current executable
	execPath, err := os.Executable()
	if err != nil {
		errStr := "Failed to retrieve the executable path"
		logger.Error().Msg(errStr)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: "",
			Error:  &errStr,
		})
	}
	// Define the script path relative to the executable directory
	scriptPath := filepath.Join(filepath.Dir(execPath), "scripts/apply.sh")
	// Run the command and capture output
	err = asyncRunCommand(scriptPath)

	if err != nil {
		errStr := fmt.Sprintf("Failed to apply resources: %s", err.Error())
		logger.Error().Msg(errStr)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}

	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// destroyResourceHandler godoc
//
//	@ID 			destroyResourceHandler
//	@Summary		Destroy Resources
//	@Description	Execute the destroy.sh script to destroy resources.
//	@Tags			[Service]
//	@Produce		json
//	@Success		200	{object}	models.BasicResponse	"Successfully destroyed resources"
//	@Failure		500	{object}	models.BasicResponse	"Failed to destroy resources"
//	@Router			/service/destroy [delete]
func (tc *TaskController) DestroyResourceHandler(ctx echo.Context) error {
	start := time.Now()
	logger, logstrings := pageLogInit(ctx, "destroy-resource", "Destroy an Data Resource", start)
	// Get the directory of the current executable
	execPath, err := os.Executable()
	if err != nil {
		errStr := "Failed to retrieve the executable path"
		logger.Error().Msg(errStr)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: "",
			Error:  &errStr,
		})
	}
	// Define the script path relative to the executable directory
	scriptPath := filepath.Join(filepath.Dir(execPath), "scripts/destroy.sh")
	// Run the command and capture output
	err = asyncRunCommand(scriptPath)

	if err != nil {
		errStr := fmt.Sprintf("Failed to destroy resources: %s", err.Error())
		logger.Error().Msg(errStr)
		return ctx.JSON(http.StatusInternalServerError, models.BasicResponse{
			Result: logstrings.String(),
			Error:  &errStr,
		})
	}

	return ctx.JSON(http.StatusOK, models.BasicResponse{
		Result: logstrings.String(),
		Error:  nil,
	})
}

// asyncRunCommand starts the command asynchronously and returns an error if it fails to start
func asyncRunCommand(scriptPath string) error {
	cmd := exec.Command("bash", scriptPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// Start the command asynchronously
	if err := cmd.Start(); err != nil {
		log.Error().Err(err).Msgf("Failed to start script %s", scriptPath)
		return fmt.Errorf("failed to start script %s: %w", scriptPath, err)
	}

	// Run in a separate goroutine to avoid blocking
	go func() {
		// Wait for the command to complete
		if err := cmd.Wait(); err != nil {
			log.Printf("Script %s finished with error: %v", scriptPath, err)
		} else {
			log.Printf("Script %s executed successfully", scriptPath)
		}
	}()

	return nil
}

// asyncRunCommandWait captures output and logs it to zerolog
func asyncRunCommandWait(scriptPath string) (string, error) {
	cmd := exec.Command("bash", scriptPath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stdout: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("failed to get stderr: %w", err)
	}

	if err := cmd.Start(); err != nil {
		log.Error().Err(err).Msgf("Failed to start script %s", scriptPath)
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	outputChan := make(chan string)
	go func() {
		var output string
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			log.Info().Msg(line)
			output += line + "\n"
		}
		scanner = bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Text()
			log.Error().Msg(line)
			output += line + "\n"
		}
		outputChan <- output
	}()

	if err := cmd.Wait(); err != nil {
		log.Error().Err(err).Msgf("Script %s finished with error", scriptPath)
		return <-outputChan, fmt.Errorf("command failed: %w", err)
	}

	log.Info().Msgf("Script %s executed successfully", scriptPath)
	return <-outputChan, nil
}
