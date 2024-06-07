include .env

BUILD_FLAGS=-ldflags="-s -w"
OUT_DIR=dist
OUT_PREFIX=nuvola

all:
	docker compose up -d
	pipenv install
	pipenv run tflocal apply -auto-approve
	pipenv run ../../nuvola dump --output-dir ./ --aws-endpoint-url http://localhost:4566