package main

import (
	"log"

	"github.com/primait/nuvola/cmd"
	"github.com/spf13/viper"
)

func main() {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln(err.Error())
	}
	cmd.Execute()
}
