package main

import (
	"fmt"
	"log"
	"time"

	"os"

	"github.com/primait/nuvola/controller/assess"
	"github.com/primait/nuvola/controller/dump"
	"github.com/primait/nuvola/controller/enumerate"

	"github.com/primait/nuvola/connector"

	"github.com/primait/nuvola/tools/cli/input/cmdflags"
	clioutput "github.com/primait/nuvola/tools/cli/output"
)

func main() {
	start := time.Now()
	// Parse command line flags
	cmdflags.InitFlags()

	switch os.Args[1] {
	case "dump":
		cloudConnector, err := connector.NewCloudConnector(cmdflags.AWS_PROFILE, cmdflags.AWS_ENDPOINT_URL)
		if err != nil {
			log.Fatalf(err.Error())
		}
		if cmdflags.DUMP_ONLY {
			dump.DumpData(nil, cloudConnector)
		} else {
			clioutput.PrintRed("Flushing Neo4j database")
			storageConnector := connector.NewStorageConnector().FlushAll()
			dump.DumpData(storageConnector, cloudConnector)
		}
		dump.SaveResults(cmdflags.AWS_PROFILE, cmdflags.OUTPUT_DIR, cmdflags.OUTPUT_FORMAT)
	case "assess":
		connector.SetActions()
		storageConnector := connector.NewStorageConnector()
		// NO_IMPORT is false so import data from INPUT_FILE first
		if cmdflags.IMPORT_FILE != "" && !cmdflags.NO_IMPORT {
			clioutput.PrintRed("Flushing all database")
			clioutput.PrintGreen(fmt.Sprintf("Importing %s", cmdflags.IMPORT_FILE))
			assess.ImportZipFile(storageConnector, cmdflags.IMPORT_FILE)
		}

		assess.Assess(storageConnector, "./controller/assess/rules/")
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
