package logging

import "github.com/fatih/color"

func (lm *logManager) PrintColored(s string, c color.Attribute) {
	_, err := color.New(c).Println(s)
	if err != nil {
		lm.Error("Error on printing colored string", "err", err)
	}
}

func (lm *logManager) PrintRed(s string) {
	lm.PrintColored(s, color.FgHiRed)
}

func (lm *logManager) PrintGreen(s string) {
	lm.PrintColored(s, color.FgHiGreen)
}

func (lm *logManager) PrintDarkGreen(s string) {
	lm.PrintColored(s, color.FgGreen)
}
