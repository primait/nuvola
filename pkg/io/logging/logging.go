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
}

type logManager struct {
	logger *log.Logger
}

var logger *logManager
var once sync.Once

func GetLogManager() LogManager {
	once.Do(func() {
		logger = &logManager{
			logger: log.NewWithOptions(os.Stdout, log.Options{
				CallerOffset:    1,
				Level:           log.InfoLevel,
				ReportCaller:    true,
				ReportTimestamp: true,
				TimeFormat:      time.RFC1123,
			}),
		}
	})

	return *logger
}

func (lm logManager) SetVerboseLevel() {
	lm.logger.SetLevel(log.InfoLevel)
}

func (lm logManager) SetDebugLevel() {
	lm.logger.SetLevel(log.DebugLevel)
}

func (lm logManager) Debug(message interface{}, keyvals ...interface{}) {
	lm.logger.Debug(message, keyvals...)
}

func (lm logManager) Info(message interface{}, keyvals ...interface{}) {
	lm.logger.Info(message, keyvals...)
}

func (lm logManager) Warn(message interface{}, keyvals ...interface{}) {
	lm.logger.Warn(message, keyvals...)
}

func (lm logManager) Error(message interface{}, keyvals ...interface{}) {
	lm.logger.Error(message, keyvals...)
	os.Exit(1)
}

var INDENT_SPACES int = 4

func PrettyJSON(s interface{}) (data []byte) {
	data, err := json.MarshalIndent(s, "", strings.Repeat(" ", INDENT_SPACES))
	if err != nil {
		if _, ok := err.(*json.UnsupportedTypeError); ok {
			return []byte("Tried to Marshal Invalid Type")
		}
		return []byte("Struct does not exist")
	}
	return
}

func JSON(s interface{}) (data []byte) {
	data, err := json.Marshal(s)
	if err != nil {
		if _, ok := err.(*json.UnsupportedTypeError); ok {
			return []byte("Tried to Marshal Invalid Type")
		}
		return []byte("Struct does not exist")
	}
	return
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
