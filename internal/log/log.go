package log

import (
	"fmt"
	"log"
	"log/syslog"
	"os"

	"github.com/pkg/errors"
)

const (
	initFailed  = "failed to initialize syslog logger"
	syslogFlags = log.Lshortfile
	normalFlags = log.LUTC | log.Ldate | log.Ltime | log.Lshortfile
)

var (
	infolog = log.New(os.Stdout, "INFO: ", normalFlags)
	errlog  = log.New(os.Stderr, "ERROR: ", normalFlags)
)

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

// Info prints a message to syslog with level LOG_NOTICE
func Info(msg string, params ...interface{}) {
	if err := infolog.Output(2, fmt.Sprintf(msg, params...)); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR writing log output: %+v", err)
	}
}

// Error prints an error to syslog with level LOG_ERR
func Error(msg string, params ...interface{}) {
	if err := errlog.Output(2, fmt.Sprintf(msg, params...)); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR writing log output: %+v", err)
	}
}

// Fatal prints an error to syslog with level LOG_ERR with Fatal
func Fatal(msg string, params ...interface{}) {
	if err := errlog.Output(2, fmt.Sprintf(msg, params...)); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR writing log output: %+v", err)
	}
	os.Exit(1)
}
