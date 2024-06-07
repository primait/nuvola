package cmd

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/primait/nuvola/pkg/connector"
	"github.com/primait/nuvola/tools/filesystem/files"
	"github.com/primait/nuvola/tools/filesystem/zip"
	"github.com/spf13/cobra"
)

var (
	AWSResults = map[string]interface{}{
		"Whoami":           nil,
		"CredentialReport": nil,
		"Groups":           nil,
		"Users":            nil,
		"Roles":            nil,
		"Buckets":          nil,
		"EC2s":             nil,
		"VPCs":             nil,
		"Lambdas":          nil,
		"RDS":              nil,
		"DynamoDBs":        nil,
		"RedshiftDBs":      nil,
	}
	dumpCmd = &cobra.Command{
		Use:   "dump",
		Short: "Dump AWS resources and policies information and store them in Neo4j",
		Run:   runDumpCmd,
	}
)

func runDumpCmd(cmd *cobra.Command, args []string) {
	startTime := time.Now()

	if cmd.Flags().Changed(flagVerbose) {
		logger.SetVerboseLevel()
	}
	if cmd.Flags().Changed(flagDebug) {
		logger.SetDebugLevel()
	}

	cloudConnector, err := connector.NewCloudConnector(awsProfile, awsEndpointUrl)
	if err != nil {
		logger.Error("Failed to create cloud connector", "err", err)
		return
	}

	if dumpOnly {
		dumpData(nil, cloudConnector)
	} else {
		storageConnector := connector.NewStorageConnector().FlushAll()
		dumpData(storageConnector, cloudConnector)
	}

	saveResults(awsProfile, outputDirectory, outputFormat)
	logger.Info("Execution Time", "seconds", time.Since(startTime))
}

func dumpData(storageConnector *connector.StorageConnector, cloudConnector *connector.CloudConnector) {
	dataChan := make(chan map[string]interface{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		cloudConnector.DumpAll("aws", dataChan, &wg)
		defer close(dataChan)
	}()

	for data := range dataChan {
		processData(data, storageConnector)
	}
	wg.Wait()
}

func processData(data map[string]interface{}, storageConnector *connector.StorageConnector) {
	for key, value := range data {
		obj, err := json.Marshal(value)
		if err != nil {
			logger.Error("Error marshalling output", "err", err)
			continue
		}
		if storageConnector != nil {
			storageConnector.ImportResults(key, obj)
		}
		AWSResults[key] = value
	}
}

func saveResults(awsProfile, outputDir, outputFormat string) {
	if awsProfile == "" {
		awsProfile = "default"
	}
	if outputFormat == "zip" {
		zip.Zip(outputDir, awsProfile, AWSResults)
	}

	today := time.Now().Format("20060102")
	for key, value := range AWSResults {
		if outputFormat == "json" {
			filename := fmt.Sprintf("%s_%s.json", key, today)
			files.PrettyJSONToFile(outputDir, filename, value)
		}
	}
}

func init() {
	rootCmd.AddCommand(dumpCmd)
}
