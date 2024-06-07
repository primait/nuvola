package connector

import (
	awsconfig "github.com/primait/nuvola/pkg/connector/services/aws"
	neo4jconnector "github.com/primait/nuvola/pkg/connector/services/neo4j"
	"github.com/primait/nuvola/pkg/io/logging"
)

type StorageConnector struct {
	Client neo4jconnector.Neo4jClient
	logger logging.LogManager
}

type CloudConnector struct {
	AWSConfig *awsconfig.AWSConfig
	logger    logging.LogManager
}
