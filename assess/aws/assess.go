package assess

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"strings"

	awsConfig "nuvola/config/aws"
	clioutput "nuvola/io/cli/output"
	unzip "nuvola/io/filesystem"
	neo4jClient "nuvola/io/neo4j"
	"nuvola/io/yamler"
)

func ImportZipFile(connector *neo4jClient.Connector, zipfile string) {
	awsConfig.SetActions()
	var ordering []string = []string{
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
	var orderedFiles []*zip.File = make([]*zip.File, len(ordering))

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
			log.Fatal(err)
		}
		defer rc.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, rc) // #nosecG110
		if err != nil {
			log.Fatal(err)
		}
		connector.ImportResults(f.Name, buf.Bytes())
	}
}

func YamlRunner(connector *neo4jClient.Connector) {
	// Now assess: perform checks based on pre-defined static rules
	for _, rule := range yamler.GetFiles("./assets/rules/") {
		var c yamler.Conf
		c.GetConf(rule)
		if c.Enabled {
			query, args := connector.PrepareQuery(&c)
			results := connector.Query(query, args)

			clioutput.PrintRed("Running rule: " + rule)
			clioutput.PrintGreen("Name: " + c.Name)
			clioutput.PrintGreen("Arguments:")
			clioutput.PrintDarkGreen(connector.ArgsToQueryNeo4jBrowser(args))
			clioutput.PrintGreen("Query:")
			clioutput.PrintDarkGreen(query)
			clioutput.PrintGreen("Description: " + c.Description)

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
