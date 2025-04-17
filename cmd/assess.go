package cmd

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/primait/nuvola/pkg/connector"
	"github.com/primait/nuvola/pkg/io/logging"
	"github.com/primait/nuvola/tools/filesystem/files"
	unzip "github.com/primait/nuvola/tools/filesystem/zip"
	"github.com/primait/nuvola/tools/yamler"
	"github.com/spf13/cobra"
)

var assessCmd = &cobra.Command{
	Use:   "assess",
	Short: "Execute assessment queries against data loaded in Neo4J",
	Run:   runAssessCmd,
}

func runAssessCmd(cmd *cobra.Command, args []string) {
	if cmd.Flags().Changed(flagVerbose) {
		logger.SetVerboseLevel()
	}
	if cmd.Flags().Changed(flagDebug) {
		logger.SetDebugLevel()
	}

	storageConnector := connector.NewStorageConnector()
	if importFile != "" {
		logger.Debug(fmt.Sprintf("Importing %s", importFile))
		importZipFile(storageConnector, importFile)
		logger.Debug(fmt.Sprintf("Imported %s", importFile))
	}

	assess(storageConnector, "./assets/rules/")
}

func importZipFile(connector *connector.StorageConnector, zipfile string) {
	connector.FlushAll()
	ordering := []string{
		"Groups", "Users", "Roles", "Buckets", "EC2s", "VPCs", "Lambdas", "RDS", "DynamoDBs", "RedshiftDBs",
	}
	orderedFiles := make([]*zip.File, len(ordering))

	r := unzip.UnzipInMemory(zipfile)
	defer func() {
		if err := r.Close(); err != nil {
			logger.Error("failed to close r: %v", err)
		}
	}()

	for _, f := range r.File {
		for ord := range ordering {
			if strings.HasPrefix(f.Name, ordering[ord]) {
				orderedFiles[ord] = f
				break
			}
		}
	}

	for _, f := range orderedFiles {
		if f == nil {
			continue
		}
		if err := processZipFile(connector, f); err != nil {
			logger.Error("Processing ZIP file", "err", err)
		}
	}
}

func processZipFile(connector *connector.StorageConnector, f *zip.File) error {
	rc, err := f.Open()
	if err != nil {
		return fmt.Errorf("opening content of ZIP: %w", err)
	}
	defer func() {
		if err := rc.Close(); err != nil {
			logger.Error("failed to close r: %v", err)
		}
	}()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, rc) // #nosecG110
	if err != nil {
		return fmt.Errorf("copying buffer from ZIP: %w", err)
	}

	connector.ImportResults(f.Name, buf.Bytes())
	return nil
}

func assess(connector *connector.StorageConnector, rulesPath string) {
	// perform checks based on pre-defined static rules
	logger := logging.GetLogManager()
	for _, rule := range files.GetFiles(rulesPath, ".ya?ml") {
		c := yamler.GetConf(rule)
		if !c.Enabled {
			continue
		}

		query, args := yamler.PrepareQuery(c)
		results := connector.Query(query, args)

		logger.PrintRed("Running rule: " + rule)
		logger.PrintGreen("Name: " + c.Name)
		logger.PrintGreen("Arguments:")
		logger.PrintDarkGreen(yamler.ArgsToQueryNeo4jBrowser(args))
		logger.PrintGreen("Query:")
		logger.PrintDarkGreen(query)
		logger.PrintGreen("Description: " + c.Description)

		for _, resultMap := range results {
			for key, value := range resultMap {
				printResults(c.Return, key, value)
			}
		}
		fmt.Print("\n")
	}
}

func printResults(returnKeys []string, key string, value interface{}) {
	for _, retValue := range returnKeys {
		if strings.HasSuffix(retValue, "*") {
			retValue = strings.TrimRight(retValue[:len(retValue)-1], "_")
		}
		if strings.HasPrefix(key, retValue) {
			fmt.Printf("%s: %v\n", key, value)
		}
	}
}

func init() {
	rootCmd.AddCommand(assessCmd)
}
