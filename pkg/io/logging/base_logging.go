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
	PrintRed(s string)
	PrintDarkGreen(s string)
	PrintGreen(s string)
	PrintColored(s string, c color.Attribute)
}

type logManager struct {
	logger *log.Logger
}

const INDENT_SPACES int = 4

var (
	instance *logManager
	once     sync.Once
)

func GetLogManager() LogManager {
	once.Do(func() {
		instance = &logManager{
			logger: log.NewWithOptions(os.Stdout, log.Options{
				CallerOffset:    1,
				Level:           log.WarnLevel,
				ReportCaller:    true,
				ReportTimestamp: true,
				TimeFormat:      time.RFC1123,
			}),
		}
	})

	return instance
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
		lm.handleJSONError(err)
	}
	return data
}

func (lm *logManager) JSON(s interface{}) []byte {
	data, err := json.Marshal(s)
	if err != nil {
		lm.handleJSONError(err)
	}
	return data
}

func (lm *logManager) handleJSONError(err error) {
	if _, ok := err.(*json.UnsupportedTypeError); ok {
		lm.Error("Tried to Marshal invalid type", "err", err)
	}
	lm.Error("Struct does not exist", "err", err)
}
