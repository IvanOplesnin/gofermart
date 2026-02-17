SHELL := /usr/bin/env bash
.SHELLFLAGS := -eu -o pipefail -c

.PHONY: run test run_memory g_up g_down

run:
	ENV_FILE=./.env ./run.sh & ./cmd/accrual/accrual_linux_amd64 -a $${ACCRUAL_RUN_ADDRESS:-localhost:8081}

test:
	go test ./... -v


run_memory:
	ENV_FILE=./.inmemory.env ./run.sh

up:
	goose -env .env up

down:
	goose -env .env down