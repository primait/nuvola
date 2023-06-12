package cmdflags

import (
	"flag"
	"log"
	nuvolaerror "nuvola/tools/error"
	"nuvola/tools/filesystem/files"
	"os"
)

var AWS_PROFILE string
var AWS_ENDPOINT_URL string
var OUTPUT_DIR string
var OUTPUT_FORMAT string
var IMPORT_FILE string
var DUMP_ONLY bool
var NO_IMPORT bool

func InitFlags() {
	var err error
	allowedFormat := map[string]bool{"zip": true, "json": true}

	dumpFlagSet := flag.NewFlagSet("dump", flag.ExitOnError)
	dumpFlagSet.StringVar(&AWS_ENDPOINT_URL, "endpoint-url", "", "AWS endpoint URL to use (e.g. for LocalStack)")
	dumpFlagSet.StringVar(&AWS_PROFILE, "profile", "", "AWS profile to use")
	dumpFlagSet.StringVar(&OUTPUT_DIR, "outputdir", "", "Output folder where the files will be saved (default: \".\")")
	dumpFlagSet.StringVar(&OUTPUT_FORMAT, "format", "zip", "Output format: ZIP or json files")
	dumpFlagSet.BoolVar(&DUMP_ONLY, "dump-only", false, "Flag to prevent loading data into Neo4j (default: \"false\")")

	assessFlagSet := flag.NewFlagSet("assess", flag.ExitOnError)
	assessFlagSet.StringVar(&IMPORT_FILE, "import", "", "Input ZIP file to load")
	assessFlagSet.BoolVar(&NO_IMPORT, "no-import", false, "Use stored data from Neo4j without import (default)")

	enumerateFlagSet := flag.NewFlagSet("enumerate", flag.ExitOnError)

	// At least a subcommand must be provided
	if len(os.Args) < 2 {
		log.Fatalln("Subcommand is required: [dump|assess]")
	}

	switch os.Args[1] {
	case "dump":
		err = dumpFlagSet.Parse(os.Args[2:])
	case "assess":
		err = assessFlagSet.Parse(os.Args[2:])
	case "enumerate":
		err = enumerateFlagSet.Parse(os.Args[2:])
	default:
		flag.PrintDefaults()
		os.Exit(1)
	}

	if err != nil {
		nuvolaerror.HandleError(err, "Cliinput - InitFlags", "Error on parsing flags")
	}

	// Check which subcommand was Parsed using the FlagSet.Parsed() function. Handle each case accordingly.
	// FlagSet.Parse() will evaluate to false if no flags were parsed (i.e. the user did not provide any flags)
	if dumpFlagSet.Parsed() {
		if !allowedFormat[OUTPUT_FORMAT] {
			log.Fatalln("Invalid output format")
		}
		OUTPUT_DIR = files.NormalizePath(OUTPUT_DIR)
	}

	if assessFlagSet.Parsed() && IMPORT_FILE != "" {
		if NO_IMPORT {
			log.Fatalln(`Select "import" OR "no-import"`)
		}
		IMPORT_FILE = files.NormalizePath(IMPORT_FILE)
	}
}
