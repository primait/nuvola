package connector

import (
	awsconfig "nuvola/connector/services/aws"
	neo4jconnector "nuvola/connector/services/neo4j"
)

type StorageConnector struct {
	Client neo4jconnector.Neo4jClient
}

type CloudConnector struct {
	AWSConfig awsconfig.AWSConfig
}
