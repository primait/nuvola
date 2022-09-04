package main

import (
	"fmt"
	"log"
	"time"

	"os"

	assess "nuvola/assess/aws"
	awsconfig "nuvola/config/aws"
	dump "nuvola/dump/aws"
	"nuvola/enumerate"
	cmdflags "nuvola/io/cli/input"
	neo4jClient "nuvola/io/neo4j"

	"github.com/joho/godotenv"
)

func main() {
	start := time.Now()
	var err error

	// Parse command line flags
	cmdflags.InitFlags()

	// Load .env
	err = godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	neo4jUrl := os.Getenv("NEO4J_URL")
	neo4jUsername := "neo4j"
	neo4jPassword := os.Getenv("PASSWORD")

	switch os.Args[1] {
	case "dump":
		connector, err := neo4jClient.Connect(neo4jUrl, neo4jUsername, neo4jPassword)
		// A Fresh dump is performed
		connector.DeleteAll()
		if err != nil {
			log.Fatalln(err)
		}
		awsConfig := awsconfig.InitAWSConfiguration(cmdflags.AWS_PROFILE)
		if cmdflags.DUMP_ONLY {
			dump.DumpData(awsConfig, nil)
		} else {
			dump.DumpData(awsConfig, &connector)
		}
	case "assess":
		connector, err := neo4jClient.Connect(neo4jUrl, neo4jUsername, neo4jPassword)
		if err != nil {
			log.Fatalln(err)
		}
		// NO_IMPORT is false so import data from INPUT_FILE first
		if cmdflags.IMPORT_FILE != "" && !cmdflags.NO_IMPORT {
			connector.DeleteAll()
			assess.ImportZipFile(&connector, cmdflags.IMPORT_FILE)
		}

		assess.YamlRunner(&connector)
	case "enumerate":
		enumerate.Enumerate()
		os.Exit(1)
	default:
		fmt.Println("You need to provide 'assess' or 'dump' flag.")
		os.Exit(1)
	}

	elapsed := time.Since(start)
	log.Printf("Elapsed: %s", elapsed)
}
