include .env

BUILD_FLAGS=-ldflags="-s -w"
OUT_DIR=dist
OUT_PREFIX=nuvola

all:
	docker-compose up -d
	pipenv run tflocal apply -auto-approve
	pipenv run ../nuvola dump --outputdir ./ --endpoint-url http://localhost:4566