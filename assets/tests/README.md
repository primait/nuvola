# Tests - nuvola

Tests are executed in a mock environment using [localstack](https://github.com/localstack/localstack). The initial configuration is performed using a Terraform project.

This folder also contains the dumped archive file from nuvola to ease the ingestion inside a Neo4j database used to execute the queries and validate the output against the exepected values.

## Usage

From the root directory of `nuvola` run `make tests`.

The Makefile will:

1. start up Neo4j docker instance
2. start up Localstack docker instance
3. load a Pipenv environment
4. using Terraform creates a mock environment
5. `nuvola` will dump the infrastructure
