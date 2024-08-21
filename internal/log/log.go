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

	"github.com/sirupsen/logrus"
)

func LogFile() {
	execPath, err := os.Executable()
	fmt.Println(execPath)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to get executable path")
	}

	// 바이너리 파일의 디렉토리 경로 가져오기
	execDir := filepath.Dir(execPath)

	// 로그 디렉토리 경로 설정
	logDir := filepath.Join(execDir, "log")

	// 로그 디렉토리 생성
	if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
		logrus.WithError(err).Fatal("Failed to create log directory")
	}

	// 로그 파일 경로 설정
	logFilePath := filepath.Join(logDir, "data-manager.log")

	// 로그 파일 열기 또는 생성
	logFile, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.FileMode(0644))
	if err != nil {
		logrus.WithError(err).Fatal("Failed to create log file")
	}
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetFormatter(&CustomTextFormatter{})
	logrus.SetOutput(io.MultiWriter(os.Stdout, logFile))
}

type CustomTextFormatter struct {
	CmdName string
	JobName string
}

func (f *CustomTextFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	timeFormatted := entry.Time.Format("2006-01-02T15:04:05-07:00")
	cn := f.CmdName
	jn := f.JobName
	if _, ok := entry.Data["cmdbName"]; ok {
		cn = entry.Data["cmdbName"].(string)
	}
	if _, ok := entry.Data["jobName"]; ok {
		jn = entry.Data["jobName"].(string)
	}
	return []byte(fmt.Sprintf("[%s] [%s] [%s] [%s] %s\n", timeFormatted, entry.Level, cn, jn, strings.ToUpper(entry.Message[:1])+entry.Message[1:])), nil
}
