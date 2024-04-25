package cmd

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/primait/nuvola/connector"
	"github.com/primait/nuvola/pkg/io/logging"
	"github.com/primait/nuvola/tools/filesystem/files"
	"github.com/primait/nuvola/tools/filesystem/zip"
	"github.com/spf13/cobra"
)

var AWSResults = map[string]interface{}{
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

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump AWS resources and policies information and store them in Neo4j",
	Run: func(cmd *cobra.Command, args []string) {
		startTime := time.Now()
		markAsRequired("aws-profile")
		if err := rootCmd.ValidateRequiredFlags(); err != nil {
			logger.Error("Required flags not provided", "err", err)
		}
		if cmd.Flags().Changed(flagVerbose) {
			logger.SetVerboseLevel()
		}
		if cmd.Flags().Changed(flagDebug) {
			logger.SetDebugLevel()
		}

		cloudConnector, err := connector.NewCloudConnector(awsProfile, awsEndpointUrl)
		if err != nil {
			logger.Error(err.Error())
		}

		if dumpOnly {
			dumpData(nil, cloudConnector)
		} else {
			storageConnector := connector.NewStorageConnector().FlushAll()
			dumpData(storageConnector, cloudConnector)
		}
		saveResults(awsProfile, outputDirectory, outputFormat)
		logger.Info("Execution Time", "seconds", time.Since(startTime))
	},
}

func dumpData(storageConnector *connector.StorageConnector, cloudConnector *connector.CloudConnector) map[string]interface{} {
	dataChan := make(chan map[string]interface{})
	go cloudConnector.DumpAll("aws", dataChan)
	for {
		a, ok := <-dataChan // receive data step by step and import it to Neo4j
		if !ok {
			break
		}
		v := reflect.ValueOf(a)
		mapKey := v.MapKeys()[0].Interface().(string)
		obj, err := json.Marshal(a[mapKey])
		if err != nil {
			logger.Error("DumpData: error marshalling output", err)
		}
		storageConnector.ImportResults(mapKey, obj)
		AWSResults[mapKey] = a[mapKey]
	}
	return AWSResults
}

func saveResults(awsProfile string, outputDir string, outputFormat string) {
	if awsProfile == "" {
		awsProfile = "default"
	}
	if outputFormat == "zip" {
		zip.Zip(outputDir, awsProfile, &AWSResults)
	}

	today := time.Now().Format("20060102")
	for key, value := range AWSResults {
		logger.Info(key, logging.PrettyJSON(value))

		if outputFormat == "json" {
			files.PrettyJSONToFile(outputDir, fmt.Sprintf("%s_%s.json", key, today), value)
		}
	}
}

func init() {
	rootCmd.AddCommand(dumpCmd)
}
