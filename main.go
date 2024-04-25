package main

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/primait/nuvola/cmd"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatalln(err.Error())
	}
	cmd.Execute()
}
