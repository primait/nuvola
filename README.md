# nuvola

[![Run Gosec](https://github.com/primait/nuvola/actions/workflows/gosec.yml/badge.svg)](https://github.com/primait/nuvola/actions/workflows/gosec.yml) [![golangci-lint](https://github.com/primait/nuvola/actions/workflows/golanci-lint.yml/badge.svg)](https://github.com/primait/nuvola/actions/workflows/golanci-lint.yml)

<p align="center">
  <img src="./assets/logo/nuvola-logo-big-light.png" height="300">
</>

nuvola (with the lowercase n) is a tool to dump and perform automatic and manual security analysis on AWS environments configurations and services using predefined, extensible and custom rules created using a simple Yaml syntax.

The general idea behind this project is to create an abstracted digital twin of a cloud platform. For a more concrete example: nuvola reflects the BloodHound traits used for Active Directory analysis but on cloud environments (at the moment only AWS).

The usage of a graph database also increases the possibility of finding different and innovative attack paths and can be used as an offline, centralised and lightweight digital twin.

## Quick Start

### Requirements

- `docker-compose` installed
- an AWS account configured to be used with `awscli` with full access to the cloud resources, better if in _ReadOnly_ mode (the policy `arn:aws:iam::aws:policy/ReadOnlyAccess` is fine)

### Setup

1. Clone the repository

```bash
git clone --depth=1 https://github.com/primait/nuvola.git; cd nuvola
```

2. Create and edit, if required, the `.env` file to set your DB username/password/URL

```bash
cp .env_example .env;
```

3. Start the Neo4j docker instance

```bash
make start
```

4. Build the tool

```bash
make build
```

### Usage

1. Firstly you need to dump all the supported AWS services configurations and load the data into the Neo4j database:

```bash
./nuvola dump -profile default_RO -outputdir ~/DumpDumpFolder -format zip
```

2. To import a previously executed dump operation into the Neo4j database:

```bash
./nuvola assess -import ~/DumpDumpFolder/nuvola-default_RO_20220901.zip
```

3. To only perform static assessments on the data loaded into the Neo4j database using the [predefined ruleset](https://github.com/primait/nuvola/tree/master/assess/rules):

```bash
./nuvola assess
```

4. Or use [Neo4j Browser](https://neo4j.com/docs/operations-manual/current/installation/neo4j-browser/) to manually explore the digital twin.

![Screenshot_20220904_185619](https://user-images.githubusercontent.com/6991986/188325663-d713d2bc-d522-4e9c-bc02-fc766f010374.png)

## About nuvola

To get started with nuvola and its database schema, check out the nuvola [Wiki](https://github.com/primait/nuvola/wiki).

No data is sent or shared with Prima Assicurazioni.

## How to contribute

- reporting bugs and issues
- reporting new improvements
- reviewing issues and pull requests
- fixing bugs and issues
- creating new rules
- improving the overall quality

## License

nuvola uses graph theory to reveal possible attack paths and security misconfigurations on cloud environments.

This program is free software: you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation, either version 3 of the License, or (at your option) any later version.

This program is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License for more details.

You should have received a copy of the GNU General Public License along with this repository and program. If not, see http://www.gnu.org/licenses/.
