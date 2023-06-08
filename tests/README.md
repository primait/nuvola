# Tests - nuvola

Tests are executed in a mock environment using [localstack](https://github.com/localstack/localstack). The initial configuration is performed using a Terraform project.

This folder also contains the dumped archive file from nuvola to ease the ingestion inside a Neo4j database used to execute the queries and validate the output against the exepected values.

## Usage

- `docker-compose up`
- `pipenv shell`
- `awslocal sts get-caller-identity` to test if the environment is ready
- `tflocal apply`
