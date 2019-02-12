package log

import (
	"fmt"
	"log"
	"log/syslog"
	"os"

	"github.com/pkg/errors"
)

// Level is a log level such a Debug or Error
type Level int

const (
	initFailed  = "failed to initialize syslog logger"
	syslogFlags = log.Lshortfile
	normalFlags = log.LUTC | log.Ldate | log.Ltime | log.Lshortfile

	// LevelDebug enables debug logging
	LevelDebug Level = iota
	// LevelError enables error logging
	LevelError Level = iota
)

var (
	infolog = log.New(os.Stdout, "INFO: ", normalFlags)
	errlog  = log.New(os.Stderr, "ERROR: ", normalFlags)

	level = LevelError
)

// SetLevel sets the log level
func SetLevel(l Level) {
	level = l
}

// InitSyslog initializes logging to syslog
func InitSyslog() (err error) {
	il, err := syslog.NewLogger(syslog.LOG_NOTICE, syslogFlags)
	if err != nil {
		return errors.Wrap(err, initFailed)
	}
	infolog = il

	el, err := syslog.NewLogger(syslog.LOG_ERR, syslogFlags)
	if err != nil {
		return errors.Wrap(err, initFailed)
	}
	errlog = el

	return nil
}

// Debug prints a debug message. If syslog is enabled then LOG_NOTICE is used
func Debug(msg string, params ...interface{}) {
	if level > LevelDebug {
		return
	}

	if err := infolog.Output(2, fmt.Sprintf(msg, params...)); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR writing log output: %+v", err)
	}
}

// Error prints an error message. If syslog is enabled then LOG_ERR is used
func Error(msg string, params ...interface{}) {
	if err := errlog.Output(2, fmt.Sprintf(msg, params...)); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR writing log output: %+v", err)
	}
}

// Fatal logs Error and exits 1
func Fatal(msg string, params ...interface{}) {
	if err := errlog.Output(2, fmt.Sprintf(msg, params...)); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR writing log output: %+v", err)
	}
	os.Exit(1)
}
