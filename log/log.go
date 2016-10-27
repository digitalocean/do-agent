// Copyright 2016 DigitalOcean
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"errors"
	stdlog "log"
	"log/syslog"
)

var (
	logLevel level

	// ErrUnrecognizedLogLevel log level not valid
	ErrUnrecognizedLogLevel = errors.New("unrecognized error level")
)

const (
	errorLevel level = iota
	infoLevel
	debugLevel
)

const (
	errorLabel = "ERROR"
	infoLabel  = "INFO"
	debugLabel = "DEBUG"
)

type level uint64

// setLevel will set the log level to one of the appropriate levels
func setLevel(level level) {
	logLevel = level
}

// setSyslogger enables the syslog writer
func setSyslogger() {
	if logwriter, err := syslog.New(syslog.LOG_NOTICE, "do-agent"); err == nil {
		stdlog.SetOutput(logwriter)
	}
}

// SetLogger sets the log level with one of the labels and sets the syslog level
func SetLogger(levelLabel string, logToSyslog bool) error {
	ll, err := toLevel(levelLabel)
	if err != nil {
		return err
	}
	setLevel(ll)
	if logToSyslog {
		setSyslogger()
	}
	return nil
}

// LogToLevel converts a log label to the corresponding level
func toLevel(label string) (level, error) {
	switch label {
	case debugLabel:
		return debugLevel, nil
	case infoLabel:
		return infoLevel, nil
	case errorLabel:
		return errorLevel, nil
	default:
		return 0, ErrUnrecognizedLogLevel
	}
}

// Debugf logs at debug level with the use of format specifiers
func Debugf(format string, args ...interface{}) {
	if logLevel >= debugLevel {
		stdlog.Printf(format, args...)
	}
}

// Infof logs at info level with the use of format specifiers
func Infof(format string, args ...interface{}) {
	if logLevel >= infoLevel {
		stdlog.Printf(format, args...)
	}
}

// Errorf logs at error level with the use of format specifiers
func Errorf(format string, args ...interface{}) {
	if logLevel >= errorLevel {
		stdlog.Printf(format, args...)
	}
}

// Fatalf logs a message with the use of format specifiers and exits
func Fatalf(format string, args ...interface{}) {
	stdlog.Fatalf(format, args...)
}

// Debug logs at debug level
func Debug(args ...interface{}) {
	if logLevel >= debugLevel {
		stdlog.Print(args...)
	}
}

// Info logs at info level
func Info(args ...interface{}) {
	if logLevel >= infoLevel {
		stdlog.Print(args...)
	}
}

// Error logs at error level
func Error(args ...interface{}) {
	if logLevel >= errorLevel {
		stdlog.Print(args...)
	}
}

// Fatal logs a message and exits
func Fatal(args ...interface{}) {
	stdlog.Fatal(args...)
}
