package log

import (
	"fmt"
	"log"
	"log/syslog"
	"os"
)

// Level is a log level such a Debug or Error
type Level int

const (
	syslogFlags = log.Llongfile
	normalFlags = log.LUTC | log.Ldate | log.Ltime | log.Llongfile

	// LevelDebug enables debug logging
	LevelDebug Level = iota
	// LevelError enables error logging
	LevelError Level = iota
)

var (
	debuglog = log.New(os.Stdout, "DEBUG: ", normalFlags)
	errlog   = log.New(os.Stderr, "ERROR: ", normalFlags)

	level = LevelError
)

// SetLevel sets the log level
func SetLevel(l Level) {
	level = l
}

// InitSyslog initializes logging to syslog
func InitSyslog() (err error) {
	dl, err := syslog.NewLogger(syslog.LOG_NOTICE, syslogFlags)
	if err != nil {
		return fmt.Errorf("InitSyslog failed to initialize debug logger: %+v", err)
	}
	debuglog = dl

	el, err := syslog.NewLogger(syslog.LOG_ERR, syslogFlags)
	if err != nil {
		return fmt.Errorf("InitSyslog failed to initialize error logger: %+v", err)
	}
	errlog = el

	return nil
}

// Debug prints a debug message. If syslog is enabled then LOG_NOTICE is used
func Debug(msg string, params ...interface{}) {
	if level > LevelDebug {
		return
	}

	if err := debuglog.Output(2, fmt.Sprintf(msg, params...)); err != nil {
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
