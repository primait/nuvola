include .env

all:
	docker compose up -d
	pipenv install
	pipenv run tflocal apply -auto-approve
	pipenv run ../../nuvola dump --output-dir ./ --aws-endpoint-url http://localhost:4566 --verbose --debug