package clioutput

import (
	"encoding/json"
	"strings"

	nuvolaerror "nuvola/tools/error"

	"github.com/fatih/color"
)

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
		nuvolaerror.HandleError(err, "Clioutput - PrintRed", "Error on printing colored string")
	}
}

func PrintGreen(s string) {
	_, err := color.New(color.FgHiGreen).Println(s)
	if err != nil {
		nuvolaerror.HandleError(err, "Clioutput - PrintRed", "Error on printing colored string")
	}
}

func PrintDarkGreen(s string) {
	_, err := color.New(color.FgGreen).Println(s)
	if err != nil {
		nuvolaerror.HandleError(err, "Clioutput - PrintRed", "Error on printing colored string")
	}
}
