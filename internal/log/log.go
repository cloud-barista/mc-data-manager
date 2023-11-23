package log

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

func LogFile() {
	logFile, err := os.OpenFile("./datamold.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, os.FileMode(0644))
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
