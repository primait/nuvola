include .env

BUILD_FLAGS=-ldflags="-s -w"
OUT_DIR=dist
OUT_PREFIX=nuvola
NEO4J_PASS ?= neo4j

define ANNOUNCE_BODY
CALL apoc.export.cypher.all("/var/lib/neo4j/import/all.cypher", {
    format: "cypher-shell",
    useOptimizations: {type: "UNWIND_BATCH", unwindBatchSize: 2000}
})
YIELD file, batches, source, format, nodes, relationships, properties, time, rows, batchSize
RETURN file, batches, source, format, nodes, relationships, properties, time, rows, batchSize;
endef

.PHONY: backup
export ANNOUNCE_BODY
backup:
	@echo "$$ANNOUNCE_BODY" | docker-compose exec -T neo4j cypher-shell -u neo4j -p ${PASSWORD} -d nuvoladb --non-interactive
	docker-compose exec -T neo4j cp /var/lib/neo4j/import/all.cypher /backup

.PHONY: check-dependencies
check-dependencies:
	@command -v docker compose >/dev/null 2>&1 || { echo >&2 "docker compose not installed"; exit 1; }
	@command -v go >/dev/null 2>&1 || { echo >&2 "Go not installed"; exit 1; }

.PHONY: check-neo4j-password
check-neo4j-password:
	@if [ -z "${NEO4J_PASS}" ]; then \
		echo "NEO4J_PASS is not set. Set it in the .env file"; \
		exit 1; \
	fi

build:
	go build ${BUILD_FLAGS} -o ${OUT_PREFIX}

compile: build
	GOOS=freebsd GOARCH=amd64 go build ${BUILD_FLAGS} -o ${OUT_DIR}/${OUT_PREFIX}-freebsd-amd64
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${OUT_DIR}/${OUT_PREFIX}-linux-amd64
	GOOS=windows GOARCH=amd64 go build ${BUILD_FLAGS} -o ${OUT_DIR}/${OUT_PREFIX}-windows-amd64
	GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -o ${OUT_DIR}/${OUT_PREFIX}-darwin-amd64
	GOOS=darwin GOARCH=arm64 go build ${BUILD_FLAGS} -o ${OUT_DIR}/${OUT_PREFIX}-darwin-arm64

.PHONY: clean
clean: stop-containers
	@rm -rf ./${OUT_DIR}
	@rm -f ./${OUT_PREFIX}

.PHONY: restore
restore:
	@cat ./backup/all.cypher | docker-compose exec -T neo4j cypher-shell -u neo4j -p ${PASSWORD} -d nuvoladb --non-interactive

.PHONY: start-containers
start-containers: check-dependencies
	@if [ ! -f ./.env ]; then\
	  cp .env_example .env;\
	fi
	@docker compose up -d

.PHONY: stop-containers
stop-containers:
	@docker compose stop
	@docker compose down -v
	@docker compose rm -fv

.PHONY: test
test: start-containers build
	cd ./assets/ && $(MAKE) -C tests all
