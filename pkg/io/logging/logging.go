package logging

import (
	"encoding/json"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/log"
	"github.com/fatih/color"
)

type LogManager interface {
	SetVerboseLevel()
	SetDebugLevel()
	Debug(message interface{}, keyvals ...interface{})
	Info(message interface{}, keyvals ...interface{})
	Warn(message interface{}, keyvals ...interface{})
	Error(message interface{}, keyvals ...interface{})
	PrettyJSON(s interface{}) []byte
	JSON(s interface{}) []byte
}

type logManager struct {
	logger *log.Logger
}

const INDENT_SPACES int = 4

var (
	logger *logManager
	once   sync.Once
)

func GetLogManager() LogManager {
	once.Do(func() {
		logger = &logManager{
			logger: log.NewWithOptions(os.Stdout, log.Options{
				CallerOffset:    1,
				Level:           log.WarnLevel,
				ReportCaller:    true,
				ReportTimestamp: true,
				TimeFormat:      time.RFC1123,
			}),
		}
	})

	return logger
}

func (lm *logManager) SetVerboseLevel() {
	lm.logger.SetLevel(log.InfoLevel)
}

func (lm *logManager) SetDebugLevel() {
	lm.logger.SetLevel(log.DebugLevel)
}

func (lm *logManager) Debug(message interface{}, keyvals ...interface{}) {
	lm.logger.Debug(message, keyvals...)
}

func (lm *logManager) Info(message interface{}, keyvals ...interface{}) {
	lm.logger.Info(message, keyvals...)
}

func (lm *logManager) Warn(message interface{}, keyvals ...interface{}) {
	lm.logger.Warn(message, keyvals...)
}

func (lm *logManager) Error(message interface{}, keyvals ...interface{}) {
	lm.logger.Error(message, keyvals...)
	os.Exit(1)
}

func (lm *logManager) PrettyJSON(s interface{}) []byte {
	data, err := json.MarshalIndent(s, "", strings.Repeat(" ", INDENT_SPACES))
	if err != nil {
		if _, ok := err.(*json.UnsupportedTypeError); ok {
			lm.Error("Tried to Marshal invalid type", "err", err)
		}
		lm.Error("Struct does not exist", "err", err)
	}
	return data
}

func (lm *logManager) JSON(s interface{}) []byte {
	data, err := json.Marshal(s)
	if err != nil {
		if _, ok := err.(*json.UnsupportedTypeError); ok {
			lm.Error("Tried to Marshal invalid type", "err", err)
		}
		lm.Error("Struct does not exist", "err", err)
	}
	return data
}

func PrintRed(s string) {
	_, err := color.New(color.FgHiRed).Println(s)
	if err != nil {
		HandleError(err, "Clioutput - PrintRed", "Error on printing colored string")
	}
}

func PrintGreen(s string) {
	_, err := color.New(color.FgHiGreen).Println(s)
	if err != nil {
		HandleError(err, "Clioutput - PrintRed", "Error on printing colored string")
	}
}

func PrintDarkGreen(s string) {
	_, err := color.New(color.FgGreen).Println(s)
	if err != nil {
		HandleError(err, "Clioutput - PrintRed", "Error on printing colored string")
	}
}
