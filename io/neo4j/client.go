package neo4j_client

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"

	neo4jConfig "nuvola/config/neo4j"
	servicesDatabase "nuvola/dump/aws/database"
	servicesEC2 "nuvola/dump/aws/ec2"
	servicesIAM "nuvola/dump/aws/iam"
	servicesLambda "nuvola/dump/aws/lambda"
	servicesS3 "nuvola/dump/aws/s3"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type Connector struct {
	neo4jConfig.Neo4jClient
}

func Connect(url, username, password string) (Connector, error) {
	client, err := neo4jConfig.Connect(url, username, password)
	if err != nil {
		log.Fatalln(err)
	}
	connector := Connector{*client}
	return connector, nil
}

func (connector *Connector) NewWriteSession() neo4j.Session {
	return connector.Driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
}

func (connector *Connector) NewReadSession() neo4j.Session {
	return connector.Driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeRead})
}

/* #nosec */
//nolint:all
func (connector *Connector) DeleteAll() {
	session := connector.NewWriteSession()
	defer session.Close()

	session.Run(`call apoc.periodic.commit("MATCH (n) WITH n LIMIT $limit DETACH DELETE n RETURN count(*)", {limit:10000})`, nil)
	session.Run("CALL apoc.schema.assert({},{})", nil)
	session.Run("CALL apoc.trigger.removeAll()", nil)

	// UNIQUE also create an index
	session.Run("CREATE CONSTRAINT IF NOT EXISTS ON (u:User) ASSERT u.Arn IS UNIQUE", nil)
	session.Run("CREATE CONSTRAINT IF NOT EXISTS ON (r:Role) ASSERT r.Arn IS UNIQUE", nil)
	session.Run("CREATE CONSTRAINT IF NOT EXISTS ON (g:Group) ASSERT g.Arn IS UNIQUE", nil)
	session.Run("CREATE CONSTRAINT IF NOT EXISTS ON (e:Ec2) ASSERT e.InstanceId IS UNIQUE", nil)
	session.Run("CREATE CONSTRAINT IF NOT EXISTS ON (b:S3) ASSERT b.Name IS UNIQUE", nil)
	session.Run("CREATE CONSTRAINT IF NOT EXISTS ON (l:Lambda) ASSERT l.FunctionConfiguration_FunctionArn IS UNIQUE", nil)
	session.Run("CREATE CONSTRAINT IF NOT EXISTS ON (v:Vpc) ASSERT v.VpcId IS UNIQUE", nil)
	session.Run("CREATE CONSTRAINT IF NOT EXISTS ON (r:Redshift) ASSERT r.ClusterIdentifier IS UNIQUE", nil)
	session.Run("CREATE CONSTRAINT IF NOT EXISTS ON (d:Dynamodb) ASSERT d.Name IS UNIQUE", nil)
	session.Run("CREATE CONSTRAINT IF NOT EXISTS ON (r:Rds) ASSERT r.DBClusterArn IS UNIQUE", nil)
	session.Run("CREATE CONSTRAINT IF NOT EXISTS ON (r:Rds) ASSERT r.DBInstanceArn IS UNIQUE", nil)

	session.Run("CREATE INDEX index_Group IF NOT EXISTS FOR (g:User) ON g.UserName", nil)

	session.Run("CREATE INDEX index_Role IF NOT EXISTS FOR (r:Role) ON r.RoleName", nil)
	session.Run("CREATE INDEX index_RoleInstanceProfileArn IF NOT EXISTS FOR (r:Role) ON r.InstanceProfileArn", nil)

	session.Run("CREATE INDEX index_Group IF NOT EXISTS FOR (g:Group) ON g.GroupName", nil)

	session.Run("CREATE INDEX index_Policy IF NOT EXISTS FOR (p:Policy) ON p.Name", nil)
	session.Run("CREATE INDEX index_PolicyId IF NOT EXISTS FOR (p:Policy) ON p.id", nil)

	session.Run("CREATE INDEX index_Action IF NOT EXISTS FOR (n:Action) ON n.Action", nil)

	session.Run("CREATE INDEX index_EC2InstanceProfiles IF NOT EXISTS FOR (e:Ec2) ON e.IamInstanceProfile_Id", nil)

	session.Run("CREATE INDEX index_VpcOwner IF NOT EXISTS FOR (v:Vpc) ON v.OwnerId", nil)
	session.Run("CREATE INDEX index_VpcType IF NOT EXISTS FOR (v:Vpc) ON v.Type", nil)

	session.Run("CREATE INDEX index_LambdaRole IF NOT EXISTS FOR (l:Lambda) ON l.Role", nil)

	session.Run("CALL db.awaitIndexes(3000)", nil)
}

func (connector *Connector) ImportResults(what string, content []byte) {
	var whoami = regexp.MustCompile(`^Whoami`)
	var credentialReport = regexp.MustCompile(`^CredentialReport`)
	var users = regexp.MustCompile(`^Users`)
	var groups = regexp.MustCompile(`^Groups`)
	var roles = regexp.MustCompile(`^Roles`)
	var buckets = regexp.MustCompile(`^Buckets`)
	var ec2s = regexp.MustCompile(`^EC2s`)
	var vpcs = regexp.MustCompile(`^VPCs`)
	var lambdas = regexp.MustCompile(`^Lambdas`)
	var rds = regexp.MustCompile(`^RDS`)
	var dynamodbs = regexp.MustCompile(`^DynamoDBs`)
	var redshiftdbs = regexp.MustCompile(`^RedshiftDBs`)

	fmt.Println(what)
	switch {
	case whoami.MatchString(what):
	case credentialReport.MatchString(what):
	case users.MatchString(what):
		contentStruct := []servicesIAM.User{}
		_ = json.Unmarshal(content, &contentStruct)
		connector.AddUsers(&contentStruct)
	case groups.MatchString(what):
		contentStruct := []servicesIAM.Group{}
		_ = json.Unmarshal(content, &contentStruct)
		connector.AddGroups(&contentStruct)
	case roles.MatchString(what):
		contentStruct := []servicesIAM.Role{}
		_ = json.Unmarshal(content, &contentStruct)
		connector.AddRoles(&contentStruct)
		connector.AddLinksToResourcesIAM()
	case buckets.MatchString(what):
		contentStruct := []servicesS3.Bucket{}
		_ = json.Unmarshal(content, &contentStruct)
		connector.AddBuckets(&contentStruct)
	case ec2s.MatchString(what):
		contentStruct := []servicesEC2.Instance{}
		_ = json.Unmarshal(content, &contentStruct)
		connector.AddEC2(&contentStruct)
	case vpcs.MatchString(what):
		contentStruct := servicesEC2.VPC{}
		_ = json.Unmarshal(content, &contentStruct)
		connector.AddVPC(&contentStruct)
	case lambdas.MatchString(what):
		contentStruct := []servicesLambda.Lambda{}
		_ = json.Unmarshal(content, &contentStruct)
		connector.AddLambda(&contentStruct)
	case rds.MatchString(what):
		contentStruct := servicesDatabase.RDS{}
		_ = json.Unmarshal(content, &contentStruct)
		connector.AddRDS(&contentStruct)
	case dynamodbs.MatchString(what):
		contentStruct := []servicesDatabase.DynamoDB{}
		_ = json.Unmarshal(content, &contentStruct)
		connector.AddDynamoDB(&contentStruct)
	case redshiftdbs.MatchString(what):
		contentStruct := []servicesDatabase.RedshiftDB{}
		_ = json.Unmarshal(content, &contentStruct)
		connector.AddRedshift(&contentStruct)
	default:
		fmt.Println("Error on importing", what)
	}
}
