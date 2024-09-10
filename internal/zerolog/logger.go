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
package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	instance *Logger
	once     sync.Once
)

type Logger struct {
	zerolog.Logger
}

type LogEntry struct {
	logger  *Logger
	level   zerolog.Level
	cmdName string
	jobName string
	message string
}

// GetInstance returns the singleton instance of Logger
func GetInstance() *Logger {
	once.Do(func() {
		instance = &Logger{
			Logger: log.Output(zerolog.ConsoleWriter{Out: os.Stderr}),
		}
		instance.setupLogger()
	})
	return instance
}

func (l *Logger) setupLogger() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatal().Msgf("Failed to get executable path: %v", err)
	}

	// Get the directory path of the binary file
	execDir := filepath.Dir(execPath)

	// Set the log directory path
	logDir := filepath.Join(execDir, "./data/var/log")

	// Create the log directory
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		log.Fatal().Msgf("Failed to create log directory: %v", err)
	}

	// Set the log file path
	logFilePath := filepath.Join(logDir, "data-manager.log")

	// Open or create the log file
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.FileMode(0644))
	if err != nil {
		log.Fatal().Msgf("Failed to create log file: %v", err)
	}

	// Set zerolog level and output
	l.Logger = l.Output(io.MultiWriter(os.Stdout, logFile)).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.DebugLevel)
}

func (l *Logger) NewLogEntry() *LogEntry {
	return &LogEntry{
		logger: l,
	}
}

func (e *LogEntry) WithLevel(level zerolog.Level) *LogEntry {
	e.level = level
	return e
}

func (e *LogEntry) WithCmdName(cmdName string) *LogEntry {
	e.cmdName = cmdName
	return e
}

func (e *LogEntry) WithJobName(jobName string) *LogEntry {
	e.jobName = jobName
	return e
}

func (e *LogEntry) WithMessage(message string) *LogEntry {
	e.message = message
	return e
}

func (e *LogEntry) logWithCustomFormat() {
	timeFormatted := time.Now().Format("2006-01-02T15:04:05-07:00")
	logEvent := e.logger.With().
		Str("time", timeFormatted).
		Str("level", e.level.String()).
		Str("cmdName", e.cmdName).
		Str("jobName", e.jobName).
		Logger()

	logEvent.Log().Msg(strings.ToUpper(e.message[:1]) + e.message[1:])
}

func Debug(cmdName, jobName string, args ...interface{}) {
	GetInstance().NewLogEntry().
		WithLevel(zerolog.DebugLevel).
		WithCmdName(cmdName).
		WithJobName(jobName).
		WithMessage(fmt.Sprint(args...)).
		logWithCustomFormat()
}

func Info(cmdName, jobName string, args ...interface{}) {
	GetInstance().NewLogEntry().
		WithLevel(zerolog.InfoLevel).
		WithCmdName(cmdName).
		WithJobName(jobName).
		WithMessage(fmt.Sprint(args...)).
		logWithCustomFormat()
}

func Warn(cmdName, jobName string, args ...interface{}) {
	GetInstance().NewLogEntry().
		WithLevel(zerolog.WarnLevel).
		WithCmdName(cmdName).
		WithJobName(jobName).
		WithMessage(fmt.Sprint(args...)).
		logWithCustomFormat()
}

func Error(cmdName, jobName string, args ...interface{}) {
	GetInstance().NewLogEntry().
		WithLevel(zerolog.ErrorLevel).
		WithCmdName(cmdName).
		WithJobName(jobName).
		WithMessage(fmt.Sprint(args...)).
		logWithCustomFormat()
}

func Fatal(cmdName, jobName string, args ...interface{}) {
	GetInstance().NewLogEntry().
		WithLevel(zerolog.FatalLevel).
		WithCmdName(cmdName).
		WithJobName(jobName).
		WithMessage(fmt.Sprint(args...)).
		logWithCustomFormat()
}
