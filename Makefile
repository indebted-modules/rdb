# Module
GO_MOD = $(shell head -n 1 go.mod|sed "s/^module //g")
MOD_NAME = $(shell basename $(GO_MOD))

# Database
DB_BASE_URL = postgres://indebted:indebted@localhost
export DB_URL = $(DB_BASE_URL)/$(MOD_NAME)?sslmode=disable
CREATE_DB = 'create database "$(MOD_NAME)"'
DROP_DB = 'drop database if exists "$(MOD_NAME)"'
DB_SCHEMA = schema/sample.sql

# Highlight
HL = @printf "\033[36m>> $1\033[0m\n"

default: help

help:
	@echo "Usage: make <TARGET>\n\nTargets:"
	@grep -E "^[\. a-zA-Z_-]+:.*?## .*$$" $(MAKEFILE_LIST) | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}' |sort

check-fmt: ## Check code formatting
	$(call HL,check-fmt)
	F=$(F) script/check-fmt.sh `find . -name "*.go"`

lint: ## Run go lint
	$(call HL,lint)
	F=$(F) script/lint.sh `find . -name "*.go"`

test: ## Run unit tests with ARGS=<go_test_args>
	$(call HL,test)
	@go test -count=1 $(ARGS) ./...

db.reset: db.drop db.create ## Reset database (drop, create)

db.create: ## Create database
	$(call HL,db.create)
	@echo ">" $(CREATE_DB)
	@psql $(DB_BASE_URL) -c $(CREATE_DB)
	@echo ">" $(DB_SCHEMA)
	@psql $(DB_URL) -f $(DB_SCHEMA)

db.drop: ## Drop database
	$(call HL,db.drop)
	@echo ">" $(DROP_DB)
	@psql $(DB_BASE_URL) -c $(DROP_DB)
