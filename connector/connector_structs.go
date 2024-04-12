package connector

import (
	awsconfig "github.com/primait/nuvola/connector/services/aws"
	neo4jconnector "github.com/primait/nuvola/connector/services/neo4j"
)

type StorageConnector struct {
	Client neo4jconnector.Neo4jClient
}

type CloudConnector struct {
	AWSConfig awsconfig.AWSConfig
}
