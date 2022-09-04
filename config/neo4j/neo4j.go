package neo4jconfig

import (
	"time"

	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
)

type Neo4jClient struct {
	Driver neo4j.Driver
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
	nc.Driver, nc.err = neo4j.NewDriver(url, neo4j.BasicAuth(username, password, ""), useConsoleLogger(logLevel), func(c *neo4j.Config) {
		c.SocketConnectTimeout = 5 * time.Second
		c.MaxConnectionLifetime = 30 * time.Minute
		// c.ConnectionAcquisitionTimeout = 5 * time.Second
	})
	if nc.err != nil {
		return &Neo4jClient{}, nc.err
	}
	return nc, nil
}
