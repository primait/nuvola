package assess

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"strings"

	connector "nuvola/connector"
	clioutput "nuvola/tools/cli/output"
	nuvolaerror "nuvola/tools/error"
	"nuvola/tools/filesystem/files"
	unzip "nuvola/tools/filesystem/zip"
	"nuvola/tools/yamler"
)

func ImportZipFile(connector *connector.StorageConnector, zipfile string) {
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
			nuvolaerror.HandleError(err, "Assess", "Opening content of ZIP")
		}
		defer rc.Close()

		buf := new(bytes.Buffer)
		_, err = io.Copy(buf, rc) // #nosecG110
		if err != nil {
			nuvolaerror.HandleError(err, "Assess", "Copying buffer from ZIP")
		}
		connector.ImportResults(f.Name, buf.Bytes())
	}
}

func Assess(connector *connector.StorageConnector, rulesPath string) {
	// perform checks based on pre-defined static rules
	for _, rule := range files.GetFiles(rulesPath, ".ya?ml") {
		var c = yamler.GetConf(rule)
		if c.Enabled {
			query, args := yamler.PrepareQuery(c)
			results := connector.Query(query, args)

			clioutput.PrintRed("Running rule: " + rule)
			clioutput.PrintGreen("Name: " + c.Name)
			clioutput.PrintGreen("Arguments:")
			clioutput.PrintDarkGreen(yamler.ArgsToQueryNeo4jBrowser(args))
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
