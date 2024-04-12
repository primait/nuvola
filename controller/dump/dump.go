package dump

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	"github.com/primait/nuvola/connector"

	clioutput "github.com/primait/nuvola/tools/cli/output"
	nuvolaerror "github.com/primait/nuvola/tools/error"
	"github.com/primait/nuvola/tools/filesystem/files"
	"github.com/primait/nuvola/tools/filesystem/zip"
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

func DumpData(storageConnector *connector.StorageConnector, cloudConnector *connector.CloudConnector) map[string]interface{} {
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
			nuvolaerror.HandleError(err, "DumpData", "Marshalling output")
		}
		storageConnector.ImportResults(mapKey, obj)
		AWSResults[mapKey] = a[mapKey]
	}
	return AWSResults
}

func SaveResults(awsProfile string, outputDir string, outputFormat string) {
	if awsProfile == "" {
		awsProfile = "default"
	}
	if outputFormat == "zip" {
		zip.Zip(outputDir, awsProfile, &AWSResults)
	}

	today := time.Now().Format("20060102")
	for key, value := range AWSResults {
		clioutput.PrintGreen(key + ":")
		fmt.Printf("%s\n", clioutput.PrettyJSON(value))

		if outputFormat == "json" {
			files.PrettyJSONToFile(outputDir, fmt.Sprintf("%s_%s.json", key, today), value)
		}
	}
}
