package cmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/primait/nuvola/connector"
	"github.com/primait/nuvola/pkg/io/logging"
	"github.com/primait/nuvola/tools/filesystem/files"
	unzip "github.com/primait/nuvola/tools/filesystem/zip"
	"github.com/primait/nuvola/tools/yamler"
	"github.com/spf13/cobra"
)

var assessCmd = &cobra.Command{
	Use:   "assess",
	Short: "Execute assessment queries agains data loaded in Neo4J",
	Run: func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed(flagVerbose) {
			logger.SetVerboseLevel()
		}
		if cmd.Flags().Changed(flagDebug) {
			logger.SetDebugLevel()
		}

		connector.SetActions()
		storageConnector := connector.NewStorageConnector()
		if importFile != "" && !noImport {
			logger.Info("Flushing database")
			logger.Info(fmt.Sprintf("Importing %s", importFile))
			importZipFile(storageConnector, importFile)
		}

		assess(storageConnector, "./assets/rules/")
	},
}

func importZipFile(connector *connector.StorageConnector, zipfile string) {
	connector.FlushAll()
	var ordering = []string{
		"Groups",
		"Users",
		"Roles",
		"Buckets",
		"EC2s",
		"VPCs",
		"Lambdas",
		"RDS",
		"DynamoDBs",
		"RedshiftDBs",
	}
	var orderedFiles = make([]*zip.File, len(ordering))

	r := unzip.UnzipInMemory(zipfile)
	defer r.Close()

	for _, f := range r.File {
		for ord := range ordering {
			if strings.HasPrefix(f.Name, ordering[ord]) {
				orderedFiles[ord] = f
			}
		}
	}

	for _, f := range orderedFiles {
		rc, err := f.Open()
		if err != nil {
			logging.HandleError(err, "Assess", "Opening content of ZIP")
		}
		defer func() {
			if err := rc.Close(); err != nil {
				log.Printf("error closing resource: %s", err)
			}
		}()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, rc) // #nosecG110
		if err != nil {
			logging.HandleError(err, "Assess", "Copying buffer from ZIP")
		}
		connector.ImportResults(f.Name, buf.Bytes())
	}
}

func assess(connector *connector.StorageConnector, rulesPath string) {
	// perform checks based on pre-defined static rules
	for _, rule := range files.GetFiles(rulesPath, ".ya?ml") {
		var c = yamler.GetConf(rule)
		if c.Enabled {
			query, args := yamler.PrepareQuery(c)
			results := connector.Query(query, args)

			logging.PrintRed("Running rule: " + rule)
			logging.PrintGreen("Name: " + c.Name)
			logging.PrintGreen("Arguments:")
			logging.PrintDarkGreen(yamler.ArgsToQueryNeo4jBrowser(args))
			logging.PrintGreen("Query:")
			logging.PrintDarkGreen(query)
			logging.PrintGreen("Description: " + c.Description)

			for _, resultMap := range results {
				for key, value := range resultMap {
					for _, retValue := range c.Return {
						if string(retValue[len(retValue)-1]) == "*" {
							// Return value contains a *: return all matching keys
							retValue = retValue[0 : len(retValue)-1]
							retValue = strings.TrimRight(retValue, "_")
						}
						if strings.HasPrefix(key, retValue) {
							fmt.Printf("%s: %v\n", key, value)
						}
					}
				}
			}
			fmt.Print("\n")
		}
	}
}

func init() {
	rootCmd.AddCommand(assessCmd)
}
