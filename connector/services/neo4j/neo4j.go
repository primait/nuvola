package neo4j_connector

import (
	"context"
	"fmt"
	"time"

	nuvolaerror "nuvola/tools/error"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j/dbtype"
)

type Neo4jClient struct {
	Driver neo4j.DriverWithContext
	err    error
}

var logLevel = neo4j.LogLevel(neo4j.WARNING)

var useConsoleLogger = func(level neo4j.LogLevel) func(config *neo4j.Config) {
	return func(config *neo4j.Config) {
		config.Log = neo4j.ConsoleLogger(level)
	}
}

func Connect(url, username, password string) (*Neo4jClient, error) {
	nc := &Neo4jClient{}
	nc.Driver, nc.err = neo4j.NewDriverWithContext(url, neo4j.BasicAuth(username, password, ""), useConsoleLogger(logLevel), func(c *neo4j.Config) {
		c.SocketConnectTimeout = 5 * time.Second
		c.MaxConnectionLifetime = 30 * time.Minute
		// c.ConnectionAcquisitionTimeout = 5 * time.Second
	})
	if nc.err != nil {
		return &Neo4jClient{}, nc.err
	}
	return nc, nil
}

func (nc *Neo4jClient) NewSession() neo4j.SessionWithContext {
	return nc.Driver.NewSession(context.TODO(), neo4j.SessionConfig{
		AccessMode: neo4j.AccessModeWrite})
}

/* #nosec */
//nolint:all
func (nc *Neo4jClient) DeleteAll() {
	session := nc.NewSession()
	defer session.Close(context.TODO())

	session.Run(context.TODO(), `call apoc.periodic.commit("MATCH (n) WITH n LIMIT $limit DETACH DELETE n RETURN count(*)", {limit:10000})`, nil)
	session.Run(context.TODO(), "CALL apoc.schema.assert({},{})", nil)
	session.Run(context.TODO(), "CALL apoc.trigger.removeAll()", nil)

	// UNIQUE also create an index
	session.Run(context.TODO(), "CREATE CONSTRAINT IF NOT EXISTS ON (u:User) ASSERT u.Arn IS UNIQUE", nil)
	session.Run(context.TODO(), "CREATE CONSTRAINT IF NOT EXISTS ON (r:Role) ASSERT r.Arn IS UNIQUE", nil)
	session.Run(context.TODO(), "CREATE CONSTRAINT IF NOT EXISTS ON (g:Group) ASSERT g.Arn IS UNIQUE", nil)
	session.Run(context.TODO(), "CREATE CONSTRAINT IF NOT EXISTS ON (e:Ec2) ASSERT e.InstanceId IS UNIQUE", nil)
	session.Run(context.TODO(), "CREATE CONSTRAINT IF NOT EXISTS ON (b:S3) ASSERT b.Name IS UNIQUE", nil)
	session.Run(context.TODO(), "CREATE CONSTRAINT IF NOT EXISTS ON (l:Lambda) ASSERT l.FunctionConfiguration_FunctionArn IS UNIQUE", nil)
	session.Run(context.TODO(), "CREATE CONSTRAINT IF NOT EXISTS ON (v:Vpc) ASSERT v.VpcId IS UNIQUE", nil)
	session.Run(context.TODO(), "CREATE CONSTRAINT IF NOT EXISTS ON (r:Redshift) ASSERT r.ClusterIdentifier IS UNIQUE", nil)
	session.Run(context.TODO(), "CREATE CONSTRAINT IF NOT EXISTS ON (d:Dynamodb) ASSERT d.Name IS UNIQUE", nil)
	session.Run(context.TODO(), "CREATE CONSTRAINT IF NOT EXISTS ON (r:Rds) ASSERT r.DBClusterArn IS UNIQUE", nil)
	session.Run(context.TODO(), "CREATE CONSTRAINT IF NOT EXISTS ON (r:Rds) ASSERT r.DBInstanceArn IS UNIQUE", nil)

	session.Run(context.TODO(), "CREATE INDEX index_Group IF NOT EXISTS FOR (g:User) ON g.UserName", nil)

	session.Run(context.TODO(), "CREATE INDEX index_Role IF NOT EXISTS FOR (r:Role) ON r.RoleName", nil)
	session.Run(context.TODO(), "CREATE INDEX index_RoleInstanceProfileArn IF NOT EXISTS FOR (r:Role) ON r.InstanceProfileArn", nil)

	session.Run(context.TODO(), "CREATE INDEX index_Group IF NOT EXISTS FOR (g:Group) ON g.GroupName", nil)

	session.Run(context.TODO(), "CREATE INDEX index_Policy IF NOT EXISTS FOR (p:Policy) ON p.Name", nil)
	session.Run(context.TODO(), "CREATE INDEX index_PolicyId IF NOT EXISTS FOR (p:Policy) ON p.id", nil)

	session.Run(context.TODO(), "CREATE INDEX index_Action IF NOT EXISTS FOR (n:Action) ON n.Action", nil)

	session.Run(context.TODO(), "CREATE INDEX index_EC2InstanceProfiles IF NOT EXISTS FOR (e:Ec2) ON e.IamInstanceProfile_Id", nil)

	session.Run(context.TODO(), "CREATE INDEX index_VpcOwner IF NOT EXISTS FOR (v:Vpc) ON v.OwnerId", nil)
	session.Run(context.TODO(), "CREATE INDEX index_VpcType IF NOT EXISTS FOR (v:Vpc) ON v.Type", nil)

	session.Run(context.TODO(), "CREATE INDEX index_LambdaRole IF NOT EXISTS FOR (l:Lambda) ON l.Role", nil)

	session.Run(context.TODO(), "CALL db.awaitIndexes(3000)", nil)
}

func (nc *Neo4jClient) Query(query string, arguments map[string]interface{}) []map[string]interface{} {
	session := nc.NewSession()
	defer session.Close(context.TODO())

	results, err := session.ExecuteWrite(context.TODO(), func(tx neo4j.ManagedTransaction) (any, error) {
		result, err := tx.Run(context.TODO(), query, arguments)
		if err != nil {
			return nil, err
		}

		results := make([]map[string]interface{}, 0)
		for result.Next(context.TODO()) {
			record, ok := result.Record().Get("result")
			if ok {
				nodeAttributes := record.(dbtype.Node).Props
				results = append(results, nodeAttributes)
			} else {
				// iterates through all results
				keys, ok := result.Keys()
				if ok == nil {
					for _, key := range keys {
						nodesMap, _ := result.Record().Get(key)
						nodeAttributes := nodesMap.(dbtype.Node).Props
						results = append(results, nodeAttributes)
					}
				}
			}
		}
		return results, result.Err()
	})
	if err != nil {
		nuvolaerror.HandleError(err, "Neo4j - Query", fmt.Sprintf("Error on executing query %s with %s", query, arguments))
	}

	return results.([]map[string]interface{})
}
