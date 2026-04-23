include .env

.PHONY: all
all:
	$(MAKE) -C ../../ start-containers build -e NEO4J_URL=${NEO4J_URL} -e NEO4J_AUTH=${NEO4J_AUTH}
	docker compose up -d
	poetry install
	poetry run tflocal init
	poetry run tflocal apply -auto-approve
	poetry run ../../nuvola dump --output-dir ./ --aws-endpoint-url http://localhost:4566 --verbose --debug