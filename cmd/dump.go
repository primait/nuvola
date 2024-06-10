package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"
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

	go func() {
		defer close(dataChan)
		cloudConnector.DumpAll("aws", dataChan)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		for data := range dataChan {
			processData(storageConnector, data)
		}
	}()
	wg.Wait()
}

func processData(storageConnector *connector.StorageConnector, data map[string]interface{}) {
	if len(data) == 0 {
		return
	}

	v := reflect.ValueOf(data)
	if v.Kind() != reflect.Map || v.Len() == 0 {
		logger.Error("processData: unexpected data format")
	}

	mapKey := v.MapKeys()[0].Interface().(string)
	obj, err := json.Marshal(data[mapKey])
	if err != nil {
		logger.Error("processData: error marshalling output", "err", err)
	}

	if storageConnector != nil {
		storageConnector.ImportResults(mapKey, obj)
	}
	AWSResults[mapKey] = data[mapKey]
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
