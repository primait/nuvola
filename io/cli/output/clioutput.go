package cli_output

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fatih/color"
)

var INDENT_SPACES int = 4

func PrettyJsonToFile(filePath string, fileName string, s interface{}) {
	if err := os.MkdirAll(filePath, os.FileMode(0775)); err != nil {
		// Error on creating/reading the output folder
		log.Fatalln(err)
	}

	filePath = filePath + string(filepath.Separator) + fileName
	if err := ioutil.WriteFile(filePath, PrettyJson(s), 0600); err != nil {
		// Error on writing file, something went wrong
		log.Fatalln(err)
	}
}

func PrettyJson(s interface{}) (data []byte) {
	data, err := json.MarshalIndent(s, "", strings.Repeat(" ", INDENT_SPACES))
	if err != nil {
		if _, ok := err.(*json.UnsupportedTypeError); ok {
			return []byte("Tried to Marshal Invalid Type")
		} else {
			return []byte("Struct does not exist")
		}
	}
	return
}

func Json(s interface{}) (data []byte) {
	data, err := json.Marshal(s)
	if err != nil {
		if _, ok := err.(*json.UnsupportedTypeError); ok {
			return []byte("Tried to Marshal Invalid Type")
		} else {
			return []byte("Struct does not exist")
		}
	}
	return
}

func PrintRed(s string) {
	_, err := color.New(color.FgHiRed).Println(s)
	if err != nil {
		log.Fatalln(err)
	}
}

func PrintGreen(s string) {
	_, err := color.New(color.FgHiGreen).Println(s)
	if err != nil {
		log.Fatalln(err)
	}
}

func PrintDarkGreen(s string) {
	_, err := color.New(color.FgGreen).Println(s)
	if err != nil {
		log.Fatalln(err)
	}
}
