SHELL := /usr/bin/env bash

default: fmt lint install generate

build:
	go build -v ./...

install: build
	go install -v ./...

lint:
	golangci-lint run

generate:
	cd tools; go generate ./...

fmt:
	gofmt -s -w -e .

test:
	go test -v -cover -timeout=120s -parallel=10 ./...

# Default: use tofu if available, fall back to terraform
TOFU_BIN ?= $(shell which tofu 2>/dev/null || which terraform)

testacc-up:
	@echo "Generating ephemeral test secrets..."
	@mkdir -p test
	@LAGO_RSA_PRIVATE_KEY=$$(openssl genrsa 2048 2>/dev/null | base64 -w0 2>/dev/null || openssl genrsa 2048 2>/dev/null | base64); \
	LAGO_ORG_API_KEY="lago-test-$$(openssl rand -hex 16)"; \
	SECRET_KEY_BASE=$$(openssl rand -hex 64); \
	LAGO_ENCRYPTION_PRIMARY_KEY=$$(openssl rand -hex 16); \
	LAGO_ENCRYPTION_DETERMINISTIC_KEY=$$(openssl rand -hex 16); \
	LAGO_ENCRYPTION_KEY_DERIVATION_SALT=$$(openssl rand -hex 16); \
	POSTGRES_PASSWORD=$$(openssl rand -hex 16); \
	LAGO_ORG_USER_PASSWORD=$$(openssl rand -hex 12); \
	printf 'LAGO_RSA_PRIVATE_KEY=%s\nLAGO_ORG_API_KEY=%s\nSECRET_KEY_BASE=%s\nLAGO_ENCRYPTION_PRIMARY_KEY=%s\nLAGO_ENCRYPTION_DETERMINISTIC_KEY=%s\nLAGO_ENCRYPTION_KEY_DERIVATION_SALT=%s\nPOSTGRES_PASSWORD=%s\nLAGO_ORG_USER_PASSWORD=%s\n' \
		"$$LAGO_RSA_PRIVATE_KEY" "$$LAGO_ORG_API_KEY" "$$SECRET_KEY_BASE" \
		"$$LAGO_ENCRYPTION_PRIMARY_KEY" "$$LAGO_ENCRYPTION_DETERMINISTIC_KEY" \
		"$$LAGO_ENCRYPTION_KEY_DERIVATION_SALT" "$$POSTGRES_PASSWORD" "$$LAGO_ORG_USER_PASSWORD" \
		> test/.lago-test-env; \
	env $$(cat test/.lago-test-env | xargs) docker compose -f docker-compose.test.yml up -d

testacc-down:
	@if [ -f test/.lago-test-env ]; then \
		env $$(cat test/.lago-test-env | xargs) docker compose -f docker-compose.test.yml down -v; \
	else \
		docker compose -f docker-compose.test.yml down -v; \
	fi
	rm -f test/.lago-test-env

testacc:
	@echo "Generating ephemeral test secrets..."
	@mkdir -p test
	@LAGO_RSA_PRIVATE_KEY=$$(openssl genrsa 2048 2>/dev/null | base64 -w0 2>/dev/null || openssl genrsa 2048 2>/dev/null | base64); \
	LAGO_ORG_API_KEY="lago-test-$$(openssl rand -hex 16)"; \
	SECRET_KEY_BASE=$$(openssl rand -hex 64); \
	LAGO_ENCRYPTION_PRIMARY_KEY=$$(openssl rand -hex 16); \
	LAGO_ENCRYPTION_DETERMINISTIC_KEY=$$(openssl rand -hex 16); \
	LAGO_ENCRYPTION_KEY_DERIVATION_SALT=$$(openssl rand -hex 16); \
	POSTGRES_PASSWORD=$$(openssl rand -hex 16); \
	LAGO_ORG_USER_PASSWORD=$$(openssl rand -hex 12); \
	printf 'LAGO_RSA_PRIVATE_KEY=%s\nLAGO_ORG_API_KEY=%s\nSECRET_KEY_BASE=%s\nLAGO_ENCRYPTION_PRIMARY_KEY=%s\nLAGO_ENCRYPTION_DETERMINISTIC_KEY=%s\nLAGO_ENCRYPTION_KEY_DERIVATION_SALT=%s\nPOSTGRES_PASSWORD=%s\nLAGO_ORG_USER_PASSWORD=%s\n' \
		"$$LAGO_RSA_PRIVATE_KEY" "$$LAGO_ORG_API_KEY" "$$SECRET_KEY_BASE" \
		"$$LAGO_ENCRYPTION_PRIMARY_KEY" "$$LAGO_ENCRYPTION_DETERMINISTIC_KEY" \
		"$$LAGO_ENCRYPTION_KEY_DERIVATION_SALT" "$$POSTGRES_PASSWORD" "$$LAGO_ORG_USER_PASSWORD" \
		> test/.lago-test-env; \
	if ! env $$(cat test/.lago-test-env | xargs) docker compose -f docker-compose.test.yml up -d; then \
		echo "ERROR: docker compose up failed"; \
		env $$(cat test/.lago-test-env | xargs) docker compose -f docker-compose.test.yml logs; \
		env $$(cat test/.lago-test-env | xargs) docker compose -f docker-compose.test.yml down -v; \
		rm -f test/.lago-test-env; \
		exit 1; \
	fi; \
	echo "Waiting for Lago API to become healthy..."; \
	READY=0; \
	for i in $$(seq 1 30); do \
		STATUS=$$(env $$(cat test/.lago-test-env | xargs) docker compose -f docker-compose.test.yml ps -q api \
			| xargs docker inspect --format='{{.State.Health.Status}}' 2>/dev/null || echo unknown); \
		if [ "$$STATUS" = "healthy" ]; then echo "Ready after $$i attempts"; READY=1; break; fi; \
		echo "Attempt $$i/30 — $$STATUS — sleeping 10s..."; sleep 10; \
	done; \
	if [ $$READY -ne 1 ]; then \
		echo "Timed out waiting for Lago API"; \
		env $$(cat test/.lago-test-env | xargs) docker compose -f docker-compose.test.yml logs; \
		env $$(cat test/.lago-test-env | xargs) docker compose -f docker-compose.test.yml down -v; \
		rm -f test/.lago-test-env; \
		exit 1; \
	fi; \
	LAGO_API_KEY=$$(grep LAGO_ORG_API_KEY test/.lago-test-env | cut -d= -f2); \
	TF_ACC=1 LAGO_ACC=1 \
	TF_ACC_TERRAFORM_PATH=$(TOFU_BIN) \
	TF_ACC_PROVIDER_NAMESPACE=rpcpool \
	LAGO_API_ENDPOINT=http://localhost:3000 \
	LAGO_API_KEY=$$LAGO_API_KEY \
	go test -v -cover -timeout 120m ./...; \
	TEST_EXIT=$$?; \
	env $$(cat test/.lago-test-env | xargs) docker compose -f docker-compose.test.yml down -v; \
	rm -f test/.lago-test-env; \
	exit $$TEST_EXIT

clean:
	rm -rf bin

.PHONY: fmt lint test testacc testacc-up testacc-down build install generate clean
