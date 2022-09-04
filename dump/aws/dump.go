package dump

import (
	"fmt"
	"log"
	"sync"
	"time"

	awsconfig "nuvola/config/aws"
	servicesDatabase "nuvola/dump/aws/database"
	servicesEC2 "nuvola/dump/aws/ec2"
	servicesIAM "nuvola/dump/aws/iam"
	servicesLambda "nuvola/dump/aws/lambda"
	servicesS3 "nuvola/dump/aws/s3"
	servicesSts "nuvola/dump/aws/sts"
	cmdflags "nuvola/io/cli/input"
	clioutput "nuvola/io/cli/output"
	zip "nuvola/io/filesystem"
	neo4jClient "nuvola/io/neo4j"
)

func DumpData(awsConfig awsconfig.AWSConfig, connector *neo4jClient.Connector) {
	var wg sync.WaitGroup
	if !testCredentials(awsConfig) {
		log.Fatalf("Invalid credentials or expired session!")
	}

	// Get the AWS regions dynamically
	servicesEC2.ListAndSaveRegions(awsConfig)

	whoami := servicesSts.Whoami(awsConfig)
	credentialReport := servicesIAM.GetCredentialReport(awsConfig)

	// The order is important; at least the IAM part must be performed initially
	groups := servicesIAM.ListGroups(awsConfig)
	wg.Add(1)
	go func() {
		if connector != nil {
			connector.ImportResults("Groups", clioutput.Json(groups))
		}
		wg.Done()
	}()

	users := servicesIAM.ListUsers(awsConfig, credentialReport)
	wg.Wait()
	wg.Add(1)
	go func() {
		if connector != nil {
			connector.ImportResults("Users", clioutput.Json(users))
		}
		wg.Done()
	}()

	roles := servicesIAM.ListRoles(awsConfig)
	wg.Wait()
	wg.Add(1)
	go func() {
		if connector != nil {
			connector.ImportResults("Roles", clioutput.Json(roles))
		}
		wg.Done()
	}()

	buckets := servicesS3.ListBuckets(awsConfig)
	wg.Wait()
	wg.Add(1)
	go func() {
		if connector != nil {
			connector.ImportResults("Buckets", clioutput.Json(buckets))
		}
		wg.Done()
	}()

	ec2 := servicesEC2.ListInstances(awsConfig)
	wg.Wait()
	wg.Add(1)
	go func() {
		if connector != nil {
			connector.ImportResults("EC2s", clioutput.Json(ec2))
		}
		wg.Done()
	}()

	vpcs := servicesEC2.ListVpcs(awsConfig)
	wg.Wait()
	wg.Add(1)
	go func() {
		if connector != nil {
			connector.ImportResults("VPCs", clioutput.Json(vpcs))
		}
		wg.Done()
	}()

	lambdas := servicesLambda.ListFunctions(awsConfig)
	wg.Wait()
	wg.Add(1)
	go func() {
		if connector != nil {
			connector.ImportResults("Lambdas", clioutput.Json(lambdas))
		}
		wg.Done()
	}()

	rds := servicesDatabase.ListRDS(awsConfig)
	wg.Wait()
	wg.Add(1)
	go func() {
		if connector != nil {
			connector.ImportResults("RDS", clioutput.Json(rds))
		}
		wg.Done()
	}()

	dynamoDBs := servicesDatabase.ListDynamoDBs(awsConfig)
	wg.Wait()
	wg.Add(1)
	go func() {
		if connector != nil {
			connector.ImportResults("DynamoDBs", clioutput.Json(dynamoDBs))
		}
		wg.Done()
	}()

	redshiftDBs := servicesDatabase.ListRedshiftDBs(awsConfig)
	wg.Wait()
	wg.Add(1)
	go func() {
		if connector != nil {
			connector.ImportResults("RedshiftDBs", clioutput.Json(redshiftDBs))
		}
		wg.Done()
	}()
	wg.Wait()

	AWSServices := map[string]interface{}{
		"Whoami":           whoami,
		"CredentialReport": credentialReport,
		"Users":            users,
		"Groups":           groups,
		"Roles":            roles,
		"Buckets":          buckets,
		"EC2s":             ec2,
		"VPCs":             vpcs,
		"Lambdas":          lambdas,
		"RDS":              rds,
		"DynamoDBs":        dynamoDBs,
		"RedshiftDBs":      redshiftDBs,
	}

	if cmdflags.OUTPUT_FORMAT == "zip" {
		zip.ZipIt(cmdflags.OUTPUT_DIR, awsConfig.Profile, &AWSServices)
	}

	today := time.Now().Format("20060102")
	for key, value := range AWSServices {
		clioutput.PrintGreen(key + ":")
		fmt.Printf("%s\n", clioutput.PrettyJson(value))

		if cmdflags.OUTPUT_FORMAT == "json" {
			clioutput.PrettyJsonToFile(cmdflags.OUTPUT_DIR, fmt.Sprintf("%s_%s.json", key, today), value)
		}
	}
}

func testCredentials(awsConfig awsconfig.AWSConfig) bool {
	output := servicesSts.Whoami(awsConfig)
	return output != nil
}
